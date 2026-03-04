import 'dart:async';

import 'package:agentic_repositories/app/app.dart';
import 'package:agentic_repositories/shared/logging/app_logger.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

Future<void> main() async {
  await runZonedGuarded(
    () async {
      WidgetsFlutterBinding.ensureInitialized();
      await AppLogger.instance.initialize();
      final logger = AppLogger.instance.logger;

      FlutterError.onError = (FlutterErrorDetails details) {
        logger.e(
          'Flutter framework error',
          error: details.exception,
          stackTrace: details.stack,
        );
        FlutterError.presentError(details);
      };

      logger.i('Launching desktop app');
      runApp(const ProviderScope(child: AgenticRepositoriesApp()));
    },
    (Object error, StackTrace stackTrace) {
      final logger = AppLogger.instance.logger;
      logger.e('Unhandled zone error', error: error, stackTrace: stackTrace);
    },
  );
}
