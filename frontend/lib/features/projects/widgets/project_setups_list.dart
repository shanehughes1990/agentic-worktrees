import 'package:agentic_worktrees/shared/graph/typed/control_plane.dart';
import 'package:flutter/material.dart';

class ProjectSetupsList extends StatelessWidget {
  const ProjectSetupsList({
    required this.projectSetups,
    required this.selectedProjectID,
    required this.onProjectSelected,
    this.dense = false,
    super.key,
  });

  final List<ProjectSetupConfig> projectSetups;
  final String selectedProjectID;
  final ValueChanged<ProjectSetupConfig> onProjectSelected;
  final bool dense;

  @override
  Widget build(BuildContext context) {
    if (projectSetups.isEmpty) {
      return const Padding(
        padding: EdgeInsets.fromLTRB(16, 0, 16, 12),
        child: Text('No project setups configured.'),
      );
    }

    return Column(
      children: projectSetups.map((ProjectSetupConfig setup) {
        final selected = selectedProjectID == setup.projectID;
        return ListTile(
          dense: dense,
          selected: selected,
          title: Text(setup.projectID),
          subtitle: Text('${setup.projectName}\n${setup.repositoryURL}'),
          onTap: () => onProjectSelected(setup),
        );
      }).toList(growable: false),
    );
  }
}
