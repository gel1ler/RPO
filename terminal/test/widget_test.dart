import 'package:flutter_test/flutter_test.dart';
import 'package:terminal/main.dart';

void main() {
  testWidgets('App renders', (WidgetTester tester) async {
    await tester.pumpWidget(const TerminalApp());
    expect(find.text('ЛР4 — NFC терминал'), findsOneWidget);
  });
}
