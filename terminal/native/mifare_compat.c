/*-
 * Derived from libnfc utils/mifare.c (BSD-licensed sample code).
 * See https://github.com/nfc-tools/libnfc
 */
#include "mifare_compat.h"

#include <string.h>

bool nfc_initiator_mifare_cmd(nfc_device *pnd, mifare_cmd mc, uint8_t ui8Block,
                               mifare_param *pmp) {
  uint8_t abtRx[265];
  size_t szParamLen;
  uint8_t abtCmd[265];

  abtCmd[0] = (uint8_t)mc;
  abtCmd[1] = ui8Block;

  switch (mc) {
  case MC_READ:
  case MC_STORE:
    szParamLen = 0;
    break;
  case MC_AUTH_A:
  case MC_AUTH_B:
    szParamLen = sizeof(struct mifare_param_auth);
    break;
  case MC_WRITE:
    szParamLen = sizeof(struct mifare_param_data);
    break;
  case MC_DECREMENT:
  case MC_INCREMENT:
  case MC_TRANSFER:
    szParamLen = sizeof(struct mifare_param_value);
    break;
  default:
    return false;
  }

  if (szParamLen) {
    memcpy(abtCmd + 2, (uint8_t *)pmp, szParamLen);
  }

  if (nfc_device_set_property_bool(pnd, NP_EASY_FRAMING, true) < 0) {
    return false;
  }

  int res =
      nfc_initiator_transceive_bytes(pnd, abtCmd, 2 + szParamLen, abtRx,
                                     sizeof(abtRx), -1);
  if (res < 0) {
    return false;
  }

  if (mc == MC_READ) {
    if (res == 16 || res == (16 + 2)) {
      memcpy(pmp->mpd.abtData, abtRx, 16);
    } else {
      return false;
    }
  }
  return true;
}
