/*-
 * Derived from libnfc utils/mifare.h (BSD-licensed sample code).
 * See https://github.com/nfc-tools/libnfc
 */
#ifndef RPO_MIFARE_COMPAT_H_
#define RPO_MIFARE_COMPAT_H_

#include <nfc/nfc.h>
#include <stdbool.h>
#include <stdint.h>

#pragma pack(1)

typedef enum {
  MC_AUTH_A = 0x60,
  MC_AUTH_B = 0x61,
  MC_READ = 0x30,
  MC_WRITE = 0xA0,
  MC_TRANSFER = 0xB0,
  MC_DECREMENT = 0xC0,
  MC_INCREMENT = 0xC1,
  MC_STORE = 0xC2
} mifare_cmd;

struct mifare_param_auth {
  uint8_t abtKey[6];
  uint8_t abtAuthUid[4];
};

struct mifare_param_data {
  uint8_t abtData[16];
};

struct mifare_param_value {
  uint8_t abtValue[4];
};

typedef union {
  struct mifare_param_auth mpa;
  struct mifare_param_data mpd;
  struct mifare_param_value mpv;
} mifare_param;

#pragma pack()

bool nfc_initiator_mifare_cmd(nfc_device *pnd, mifare_cmd mc, uint8_t ui8Block,
                               mifare_param *pmp);

#endif
