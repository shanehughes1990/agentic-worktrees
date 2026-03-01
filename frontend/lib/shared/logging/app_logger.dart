import 'dart:async';
import 'dart:io';

import 'package:logger/logger.dart';
import 'package:path_provider/path_provider.dart';

class AppLogger {
  AppLogger._();

  static final AppLogger instance = AppLogger._();

  Logger? _logger;
  IOSink? _sink;
  File? _logFile;

  Logger get logger {
    final existing = _logger;
    if (existing != null) {
      return existing;
    }

    final fallback = Logger(
      printer: PrettyPrinter(
        methodCount: 0,
        errorMethodCount: 8,
        lineLength: 120,
        colors: false,
        printEmojis: false,
        dateTimeFormat: DateTimeFormat.onlyTimeAndSinceStart,
      ),
      level: Level.trace,
      output: ConsoleOutput(),
    );

    _logger = fallback;
    fallback.w(
      'AppLogger used before initialize(); falling back to console output only.',
    );
    return fallback;
  }

  String? get logFilePath => _logFile?.path;

  Future<void> initialize() async {
    if (_logger != null) {
      return;
    }

    final supportDirectory = await getApplicationSupportDirectory();
    final logDirectory = Directory('${supportDirectory.path}/logs');
    if (!await logDirectory.exists()) {
      await logDirectory.create(recursive: true);
    }

    final logFile = File('${logDirectory.path}/desktop.log');
    _logFile = logFile;
    _sink = logFile.openWrite(mode: FileMode.append);

    final output = MultiOutput(<LogOutput>[
      ConsoleOutput(),
      _FileLogOutput(_sink!),
    ]);

    _logger = Logger(
      printer: PrettyPrinter(
        methodCount: 0,
        errorMethodCount: 8,
        lineLength: 120,
        colors: false,
        printEmojis: false,
        dateTimeFormat: DateTimeFormat.onlyTimeAndSinceStart,
      ),
      level: Level.trace,
      output: output,
    );

    logger.i('Logger initialized', error: {'path': logFile.path});
  }

  Future<void> dispose() async {
    await _sink?.flush();
    await _sink?.close();
  }
}

class _FileLogOutput extends LogOutput {
  _FileLogOutput(this._sink);

  final IOSink _sink;

  @override
  void output(OutputEvent event) {
    for (final line in event.lines) {
      _sink.writeln(line);
    }
  }
}
