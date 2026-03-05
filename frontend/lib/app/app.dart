import 'package:agentic_repositories/features/dashboard/dashboard_screen.dart';
import 'package:agentic_repositories/shared/config/app_config.dart';
import 'package:agentic_repositories/shared/logging/app_logger.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class AgenticRepositoriesApp extends ConsumerWidget {
  const AgenticRepositoriesApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final config = ref.watch(appConfigProvider);
    return MaterialApp(
      title: 'Agentic Repositories',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.indigo),
      ),
      home: config.when(
        data: (value) =>
            DashboardScreen(initialEndpoint: value.graphqlHttpEndpoint),
        loading: () =>
            const Scaffold(body: Center(child: CircularProgressIndicator())),
        error: (error, stack) {
          AppLogger.instance.logger.e(
            'Failed to load app config',
            error: error,
            stackTrace: stack,
          );
          return Scaffold(
            body: Center(child: Text('Failed to load app config: $error')),
          );
        },
      ),
    );
  }
}
