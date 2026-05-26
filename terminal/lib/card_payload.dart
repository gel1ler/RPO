import 'dart:convert';

class CardPayload {
  CardPayload({
    required this.v,
    required this.cardNumber,
    required this.balance,
    required this.keyId,
  });

  factory CardPayload.fromJsonMap(Map<String, dynamic> m) {
    return CardPayload(
      v: (m['v'] as num?)?.toInt() ?? 1,
      cardNumber: (m['card_number'] ?? '') as String,
      balance: (m['balance'] as num?)?.toInt() ?? 0,
      keyId: (m['key_id'] as num?)?.toInt() ?? 1,
    );
  }

  int v;
  String cardNumber;
  int balance;
  int keyId;

  factory CardPayload.parse(String raw) =>
      CardPayload.fromJsonMap(jsonDecode(raw) as Map<String, dynamic>);

  Map<String, dynamic> toJson() => {
        'v': v,
        'card_number': cardNumber,
        'balance': balance,
        'key_id': keyId,
      };

  String toJsonString() => jsonEncode(toJson());
}
