#include "nfc_card.h"
#include "mifare_compat.h"

#include <nfc/nfc.h>
#include <openssl/evp.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define CARD_WAIT_TIMEOUT_MS 5000
#define CARD_POLL_INTERVAL_MS 150

#define DATA_BLOCK_COUNT 6
#define BLOCK_SIZE 16
#define MAX_PAYLOAD (DATA_BLOCK_COUNT * BLOCK_SIZE)

static char g_last_error[256] = "";

static void set_error(const char *msg) {
  snprintf(g_last_error, sizeof(g_last_error), "%s", msg ? msg : "");
}

const char *nfc_last_error(void) { return g_last_error; }

typedef struct {
  nfc_context *ctx;
  nfc_device *dev;
} nfc_session_t;

static void session_close(nfc_session_t *s) {
  if (s->dev) {
    nfc_close(s->dev);
    s->dev = NULL;
  }
  if (s->ctx) {
    nfc_exit(s->ctx);
    s->ctx = NULL;
  }
}

static int session_open(nfc_session_t *s) {
  memset(s, 0, sizeof(*s));
  nfc_init(&s->ctx);
  if (!s->ctx) {
    set_error("nfc_init failed");
    return NFC_CARD_ERR;
  }
  s->dev = nfc_open(s->ctx, NULL);
  if (!s->dev) {
    set_error("nfc_open failed (no reader?)");
    nfc_exit(s->ctx);
    s->ctx = NULL;
    return NFC_CARD_ERR;
  }
  if (nfc_initiator_init(s->dev) < 0) {
    set_error("nfc_initiator_init failed");
    nfc_close(s->dev);
    nfc_exit(s->ctx);
    s->dev = NULL;
    s->ctx = NULL;
    return NFC_CARD_ERR;
  }
  return NFC_CARD_OK;
}

static int hex_to_bytes(const char *hex, uint8_t *out, size_t out_len) {
  if (strlen(hex) != 12 || out_len < 6) {
    set_error("mifare key must be 12 hex chars");
    return NFC_CARD_ERR;
  }
  for (size_t i = 0; i < 6; i++) {
    unsigned int b = 0;
    if (sscanf(hex + i * 2, "%2x", &b) != 1) {
      set_error("invalid hex in mifare key");
      return NFC_CARD_ERR;
    }
    out[i] = (uint8_t)b;
  }
  return NFC_CARD_OK;
}

static void derive_aes_key(const uint8_t mifare_key[6], uint8_t aes_key[16]) {
  memcpy(aes_key, mifare_key, 6);
  memcpy(aes_key + 6, mifare_key, 6);
  memcpy(aes_key + 12, mifare_key, 4); /* fills 16 bytes from 48-bit secret */
}

static int decrypt_block_aes(const uint8_t aes_key[16], const uint8_t in[BLOCK_SIZE],
                             uint8_t out[BLOCK_SIZE]) {
  EVP_CIPHER_CTX *ctx = EVP_CIPHER_CTX_new();
  if (!ctx) {
    set_error("EVP_CIPHER_CTX_new failed");
    return NFC_CARD_ERR;
  }
  int ok = 0;
  int outl = 0;
  if (EVP_DecryptInit_ex(ctx, EVP_aes_128_ecb(), NULL, aes_key, NULL) == 1 &&
      EVP_CIPHER_CTX_set_padding(ctx, 0) == 1 &&
      EVP_DecryptUpdate(ctx, out, &outl, in, BLOCK_SIZE) == 1 &&
      EVP_DecryptFinal_ex(ctx, out + outl, &outl) == 1) {
    ok = 1;
  }
  EVP_CIPHER_CTX_free(ctx);
  if (!ok) {
    set_error("decrypt_block failed");
    return NFC_CARD_ERR;
  }
  return NFC_CARD_OK;
}

static int encrypt_block_aes(const uint8_t aes_key[16], const uint8_t in[BLOCK_SIZE],
                             uint8_t out[BLOCK_SIZE]) {
  EVP_CIPHER_CTX *ctx = EVP_CIPHER_CTX_new();
  if (!ctx) {
    set_error("EVP_CIPHER_CTX_new failed");
    return NFC_CARD_ERR;
  }
  int ok = 0;
  int outl = 0;
  if (EVP_EncryptInit_ex(ctx, EVP_aes_128_ecb(), NULL, aes_key, NULL) == 1 &&
      EVP_CIPHER_CTX_set_padding(ctx, 0) == 1 &&
      EVP_EncryptUpdate(ctx, out, &outl, in, BLOCK_SIZE) == 1 &&
      EVP_EncryptFinal_ex(ctx, out + outl, &outl) == 1) {
    ok = 1;
  }
  EVP_CIPHER_CTX_free(ctx);
  if (!ok) {
    set_error("encrypt_block failed");
    return NFC_CARD_ERR;
  }
  return NFC_CARD_OK;
}

static void sleep_ms(unsigned ms) { usleep((useconds_t)ms * 1000); }

/** Ожидание карты до CARD_WAIT_TIMEOUT_MS с периодическим опросом. */
static int select_card(nfc_device *dev, nfc_target *nt) {
  const nfc_modulation nm = {.nmt = NMT_ISO14443A, .nbr = NBR_106};
  unsigned elapsed = 0;

  while (1) {
    if (nfc_initiator_select_passive_target(dev, nm, NULL, 0, nt) > 0) {
      return NFC_CARD_OK;
    }
    if (elapsed >= CARD_WAIT_TIMEOUT_MS) {
      break;
    }
    sleep_ms(CARD_POLL_INTERVAL_MS);
    elapsed += CARD_POLL_INTERVAL_MS;
  }
  set_error("no card on reader (waited 5s)");
  return NFC_CARD_ERR;
}

static void format_uid(const nfc_target *nt, char *uid_out, size_t uid_len) {
  if (!uid_out || uid_len < 8) {
    return;
  }
  uid_out[0] = '\0';
  if (nt->nti.nai.szUidLen < 4) {
    return;
  }
  const uint8_t *u = nt->nti.nai.abtUid;
  size_t n = nt->nti.nai.szUidLen;
  if (n > 4) {
    u += n - 4;
  }
  snprintf(uid_out, uid_len, "%02X%02X%02X%02X", u[0], u[1], u[2], u[3]);
}

static int mifare_auth(nfc_device *dev, const nfc_target *nt, uint8_t block,
                       const uint8_t mifare_key[6]) {
  mifare_param mp;
  memcpy(mp.mpa.abtKey, mifare_key, 6);
  if (nt->nti.nai.szUidLen >= 4) {
    memcpy(mp.mpa.abtAuthUid,
           nt->nti.nai.abtUid + nt->nti.nai.szUidLen - 4, 4);
  } else {
    set_error("unexpected uid length");
    return NFC_CARD_ERR;
  }
  if (!nfc_initiator_mifare_cmd(dev, MC_AUTH_A, block, &mp)) {
    set_error("mifare auth failed");
    return NFC_CARD_ERR;
  }
  return NFC_CARD_OK;
}

/** Data blocks: sector1 (4–6) and sector2 (8–10), Key A on sector trailer. */
static const uint8_t k_data_blocks[DATA_BLOCK_COUNT] = {4, 5, 6, 8, 9, 10};

static int rw_blocks(nfc_session_t *session, const uint8_t mifare_key[6],
                     int write, const uint8_t plain[MAX_PAYLOAD],
                     uint8_t *plain_out, const nfc_target *nt) {
  uint8_t aes_key[16];
  derive_aes_key(mifare_key, aes_key);

  for (int i = 0; i < DATA_BLOCK_COUNT; i++) {
    uint8_t block = k_data_blocks[i];
    if (mifare_auth(session->dev, nt, block, mifare_key) != NFC_CARD_OK) {
      return NFC_CARD_ERR;
    }

    mifare_param mp;
    uint8_t wire[BLOCK_SIZE];
    if (write) {
      const uint8_t *src = plain + i * BLOCK_SIZE;
      if (encrypt_block_aes(aes_key, src, wire) != NFC_CARD_OK) {
        return NFC_CARD_ERR;
      }
      memcpy(mp.mpd.abtData, wire, BLOCK_SIZE);
      if (!nfc_initiator_mifare_cmd(session->dev, MC_WRITE, block, &mp)) {
        set_error("mifare write failed");
        return NFC_CARD_ERR;
      }
    } else {
      if (!nfc_initiator_mifare_cmd(session->dev, MC_READ, block, &mp)) {
        set_error("mifare read failed");
        return NFC_CARD_ERR;
      }
      uint8_t *dst = plain_out + i * BLOCK_SIZE;
      if (decrypt_block_aes(aes_key, mp.mpd.abtData, dst) != NFC_CARD_OK) {
        return NFC_CARD_ERR;
      }
    }
  }
  return NFC_CARD_OK;
}

int nfc_reader_present(void) {
  nfc_context *ctx = NULL;
  nfc_init(&ctx);
  if (!ctx) {
    return 0;
  }
  nfc_connstring connstrings[8];
  size_t n = nfc_list_devices(ctx, connstrings, 8);
  nfc_exit(ctx);
  return n > 0 ? 1 : 0;
}

int nfc_read_card(const char *mifare_key_hex, char *uid_out, size_t uid_len,
                  char *json_out, size_t json_len) {
  uint8_t mifare_key[6];
  if (hex_to_bytes(mifare_key_hex, mifare_key, sizeof(mifare_key)) !=
      NFC_CARD_OK) {
    return NFC_CARD_ERR;
  }

  nfc_session_t s;
  if (session_open(&s) != NFC_CARD_OK) {
    return NFC_CARD_ERR;
  }

  nfc_target nt;
  memset(&nt, 0, sizeof(nt));
  if (select_card(s.dev, &nt) != NFC_CARD_OK) {
    session_close(&s);
    return NFC_CARD_ERR;
  }
  if (uid_out && uid_len > 0) {
    format_uid(&nt, uid_out, uid_len);
  }

  uint8_t plain[MAX_PAYLOAD];
  memset(plain, 0, sizeof(plain));
  if (rw_blocks(&s, mifare_key, 0, plain, plain, &nt) != NFC_CARD_OK) {
    session_close(&s);
    return NFC_CARD_ERR;
  }
  session_close(&s);

  plain[MAX_PAYLOAD - 1] = '\0';
  size_t n = strnlen((char *)plain, MAX_PAYLOAD);
  if (json_len == 0 || n >= json_len) {
    set_error("json buffer too small");
    return NFC_CARD_ERR;
  }
  memcpy(json_out, plain, n + 1);
  return NFC_CARD_OK;
}

int nfc_write_card(const char *mifare_key_hex, const char *json_in) {
  uint8_t mifare_key[6];
  if (hex_to_bytes(mifare_key_hex, mifare_key, sizeof(mifare_key)) !=
      NFC_CARD_OK) {
    return NFC_CARD_ERR;
  }
  if (!json_in) {
    set_error("json is null");
    return NFC_CARD_ERR;
  }

  nfc_session_t s;
  if (session_open(&s) != NFC_CARD_OK) {
    return NFC_CARD_ERR;
  }

  nfc_target nt;
  memset(&nt, 0, sizeof(nt));
  if (select_card(s.dev, &nt) != NFC_CARD_OK) {
    session_close(&s);
    return NFC_CARD_ERR;
  }

  uint8_t plain[MAX_PAYLOAD];
  memset(plain, 0, sizeof(plain));
  size_t n = strlen(json_in);
  if (n >= MAX_PAYLOAD) {
    set_error("json too large for card");
    session_close(&s);
    return NFC_CARD_ERR;
  }
  memcpy(plain, json_in, n);

  if (rw_blocks(&s, mifare_key, 1, plain, NULL, &nt) != NFC_CARD_OK) {
    session_close(&s);
    return NFC_CARD_ERR;
  }
  session_close(&s);
  return NFC_CARD_OK;
}
