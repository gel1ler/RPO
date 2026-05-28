#ifndef NFC_CARD_H
#define NFC_CARD_H

#include <stddef.h>

#define NFC_CARD_OK 0
#define NFC_CARD_ERR -1

#ifdef __cplusplus
extern "C" {
#endif

int nfc_reader_present(void);
int nfc_read_uid(char *uid_out, size_t uid_len);
int nfc_read_card(const char *mifare_key_hex, char *uid_out, size_t uid_len,
                  char *json_out, size_t json_len);
int nfc_write_card(const char *mifare_key_hex, const char *json_in);
const char *nfc_last_error(void);

#ifdef __cplusplus
}
#endif

#endif
