class Input$SupervisorCorrelationInput {
  factory Input$SupervisorCorrelationInput({
    required String runID,
    required String taskID,
    required String jobID,
  }) => Input$SupervisorCorrelationInput._({
    r'runID': runID,
    r'taskID': taskID,
    r'jobID': jobID,
  });

  Input$SupervisorCorrelationInput._(this._$data);

  factory Input$SupervisorCorrelationInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$runID = data['runID'];
    result$data['runID'] = (l$runID as String);
    final l$taskID = data['taskID'];
    result$data['taskID'] = (l$taskID as String);
    final l$jobID = data['jobID'];
    result$data['jobID'] = (l$jobID as String);
    return Input$SupervisorCorrelationInput._(result$data);
  }

  Map<String, dynamic> _$data;

  String get runID => (_$data['runID'] as String);

  String get taskID => (_$data['taskID'] as String);

  String get jobID => (_$data['jobID'] as String);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$runID = runID;
    result$data['runID'] = l$runID;
    final l$taskID = taskID;
    result$data['taskID'] = l$taskID;
    final l$jobID = jobID;
    result$data['jobID'] = l$jobID;
    return result$data;
  }

  CopyWith$Input$SupervisorCorrelationInput<Input$SupervisorCorrelationInput>
  get copyWith => CopyWith$Input$SupervisorCorrelationInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$SupervisorCorrelationInput ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$runID = runID;
    final lOther$runID = other.runID;
    if (l$runID != lOther$runID) {
      return false;
    }
    final l$taskID = taskID;
    final lOther$taskID = other.taskID;
    if (l$taskID != lOther$taskID) {
      return false;
    }
    final l$jobID = jobID;
    final lOther$jobID = other.jobID;
    if (l$jobID != lOther$jobID) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$runID = runID;
    final l$taskID = taskID;
    final l$jobID = jobID;
    return Object.hashAll([l$runID, l$taskID, l$jobID]);
  }
}

abstract class CopyWith$Input$SupervisorCorrelationInput<TRes> {
  factory CopyWith$Input$SupervisorCorrelationInput(
    Input$SupervisorCorrelationInput instance,
    TRes Function(Input$SupervisorCorrelationInput) then,
  ) = _CopyWithImpl$Input$SupervisorCorrelationInput;

  factory CopyWith$Input$SupervisorCorrelationInput.stub(TRes res) =
      _CopyWithStubImpl$Input$SupervisorCorrelationInput;

  TRes call({String? runID, String? taskID, String? jobID});
}

class _CopyWithImpl$Input$SupervisorCorrelationInput<TRes>
    implements CopyWith$Input$SupervisorCorrelationInput<TRes> {
  _CopyWithImpl$Input$SupervisorCorrelationInput(this._instance, this._then);

  final Input$SupervisorCorrelationInput _instance;

  final TRes Function(Input$SupervisorCorrelationInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? runID = _undefined,
    Object? taskID = _undefined,
    Object? jobID = _undefined,
  }) => _then(
    Input$SupervisorCorrelationInput._({
      ..._instance._$data,
      if (runID != _undefined && runID != null) 'runID': (runID as String),
      if (taskID != _undefined && taskID != null) 'taskID': (taskID as String),
      if (jobID != _undefined && jobID != null) 'jobID': (jobID as String),
    }),
  );
}

class _CopyWithStubImpl$Input$SupervisorCorrelationInput<TRes>
    implements CopyWith$Input$SupervisorCorrelationInput<TRes> {
  _CopyWithStubImpl$Input$SupervisorCorrelationInput(this._res);

  TRes _res;

  call({String? runID, String? taskID, String? jobID}) => _res;
}

enum Enum$SupervisorSignalType {
  JOB_ADMITTED,
  EXECUTION_PROGRESSED,
  EXECUTION_FAILED,
  EXECUTION_SUCCEEDED,
  CHECKPOINT_SAVED,
  TRACKER_ATTENTION_NEEDED,
  TRACKER_ATTENTION_CLEARED,
  SCM_ATTENTION_NEEDED,
  SCM_ATTENTION_CLEARED,
  PR_CONFLICT_DETECTED,
  PR_REVIEW_CHANGES_REQUESTED,
  PR_CHECKS_FAILED,
  PR_CHECKS_PASSED,
  PR_MERGE_REQUESTED,
  ISSUE_OPENED,
  ISSUE_APPROVED,
  MANUAL_OVERRIDE,
  $unknown;

  factory Enum$SupervisorSignalType.fromJson(String value) =>
      fromJson$Enum$SupervisorSignalType(value);

  String toJson() => toJson$Enum$SupervisorSignalType(this);
}

String toJson$Enum$SupervisorSignalType(Enum$SupervisorSignalType e) {
  switch (e) {
    case Enum$SupervisorSignalType.JOB_ADMITTED:
      return r'JOB_ADMITTED';
    case Enum$SupervisorSignalType.EXECUTION_PROGRESSED:
      return r'EXECUTION_PROGRESSED';
    case Enum$SupervisorSignalType.EXECUTION_FAILED:
      return r'EXECUTION_FAILED';
    case Enum$SupervisorSignalType.EXECUTION_SUCCEEDED:
      return r'EXECUTION_SUCCEEDED';
    case Enum$SupervisorSignalType.CHECKPOINT_SAVED:
      return r'CHECKPOINT_SAVED';
    case Enum$SupervisorSignalType.TRACKER_ATTENTION_NEEDED:
      return r'TRACKER_ATTENTION_NEEDED';
    case Enum$SupervisorSignalType.TRACKER_ATTENTION_CLEARED:
      return r'TRACKER_ATTENTION_CLEARED';
    case Enum$SupervisorSignalType.SCM_ATTENTION_NEEDED:
      return r'SCM_ATTENTION_NEEDED';
    case Enum$SupervisorSignalType.SCM_ATTENTION_CLEARED:
      return r'SCM_ATTENTION_CLEARED';
    case Enum$SupervisorSignalType.PR_CONFLICT_DETECTED:
      return r'PR_CONFLICT_DETECTED';
    case Enum$SupervisorSignalType.PR_REVIEW_CHANGES_REQUESTED:
      return r'PR_REVIEW_CHANGES_REQUESTED';
    case Enum$SupervisorSignalType.PR_CHECKS_FAILED:
      return r'PR_CHECKS_FAILED';
    case Enum$SupervisorSignalType.PR_CHECKS_PASSED:
      return r'PR_CHECKS_PASSED';
    case Enum$SupervisorSignalType.PR_MERGE_REQUESTED:
      return r'PR_MERGE_REQUESTED';
    case Enum$SupervisorSignalType.ISSUE_OPENED:
      return r'ISSUE_OPENED';
    case Enum$SupervisorSignalType.ISSUE_APPROVED:
      return r'ISSUE_APPROVED';
    case Enum$SupervisorSignalType.MANUAL_OVERRIDE:
      return r'MANUAL_OVERRIDE';
    case Enum$SupervisorSignalType.$unknown:
      return r'$unknown';
  }
}

Enum$SupervisorSignalType fromJson$Enum$SupervisorSignalType(String value) {
  switch (value) {
    case r'JOB_ADMITTED':
      return Enum$SupervisorSignalType.JOB_ADMITTED;
    case r'EXECUTION_PROGRESSED':
      return Enum$SupervisorSignalType.EXECUTION_PROGRESSED;
    case r'EXECUTION_FAILED':
      return Enum$SupervisorSignalType.EXECUTION_FAILED;
    case r'EXECUTION_SUCCEEDED':
      return Enum$SupervisorSignalType.EXECUTION_SUCCEEDED;
    case r'CHECKPOINT_SAVED':
      return Enum$SupervisorSignalType.CHECKPOINT_SAVED;
    case r'TRACKER_ATTENTION_NEEDED':
      return Enum$SupervisorSignalType.TRACKER_ATTENTION_NEEDED;
    case r'TRACKER_ATTENTION_CLEARED':
      return Enum$SupervisorSignalType.TRACKER_ATTENTION_CLEARED;
    case r'SCM_ATTENTION_NEEDED':
      return Enum$SupervisorSignalType.SCM_ATTENTION_NEEDED;
    case r'SCM_ATTENTION_CLEARED':
      return Enum$SupervisorSignalType.SCM_ATTENTION_CLEARED;
    case r'PR_CONFLICT_DETECTED':
      return Enum$SupervisorSignalType.PR_CONFLICT_DETECTED;
    case r'PR_REVIEW_CHANGES_REQUESTED':
      return Enum$SupervisorSignalType.PR_REVIEW_CHANGES_REQUESTED;
    case r'PR_CHECKS_FAILED':
      return Enum$SupervisorSignalType.PR_CHECKS_FAILED;
    case r'PR_CHECKS_PASSED':
      return Enum$SupervisorSignalType.PR_CHECKS_PASSED;
    case r'PR_MERGE_REQUESTED':
      return Enum$SupervisorSignalType.PR_MERGE_REQUESTED;
    case r'ISSUE_OPENED':
      return Enum$SupervisorSignalType.ISSUE_OPENED;
    case r'ISSUE_APPROVED':
      return Enum$SupervisorSignalType.ISSUE_APPROVED;
    case r'MANUAL_OVERRIDE':
      return Enum$SupervisorSignalType.MANUAL_OVERRIDE;
    default:
      return Enum$SupervisorSignalType.$unknown;
  }
}

enum Enum$SupervisorState {
  IDLE,
  EXECUTING,
  REVIEWING,
  REWORK,
  MERGE_READY,
  BLOCKED,
  ESCALATED,
  MERGED,
  REFUSED,
  COMPLETED,
  $unknown;

  factory Enum$SupervisorState.fromJson(String value) =>
      fromJson$Enum$SupervisorState(value);

  String toJson() => toJson$Enum$SupervisorState(this);
}

String toJson$Enum$SupervisorState(Enum$SupervisorState e) {
  switch (e) {
    case Enum$SupervisorState.IDLE:
      return r'IDLE';
    case Enum$SupervisorState.EXECUTING:
      return r'EXECUTING';
    case Enum$SupervisorState.REVIEWING:
      return r'REVIEWING';
    case Enum$SupervisorState.REWORK:
      return r'REWORK';
    case Enum$SupervisorState.MERGE_READY:
      return r'MERGE_READY';
    case Enum$SupervisorState.BLOCKED:
      return r'BLOCKED';
    case Enum$SupervisorState.ESCALATED:
      return r'ESCALATED';
    case Enum$SupervisorState.MERGED:
      return r'MERGED';
    case Enum$SupervisorState.REFUSED:
      return r'REFUSED';
    case Enum$SupervisorState.COMPLETED:
      return r'COMPLETED';
    case Enum$SupervisorState.$unknown:
      return r'$unknown';
  }
}

Enum$SupervisorState fromJson$Enum$SupervisorState(String value) {
  switch (value) {
    case r'IDLE':
      return Enum$SupervisorState.IDLE;
    case r'EXECUTING':
      return Enum$SupervisorState.EXECUTING;
    case r'REVIEWING':
      return Enum$SupervisorState.REVIEWING;
    case r'REWORK':
      return Enum$SupervisorState.REWORK;
    case r'MERGE_READY':
      return Enum$SupervisorState.MERGE_READY;
    case r'BLOCKED':
      return Enum$SupervisorState.BLOCKED;
    case r'ESCALATED':
      return Enum$SupervisorState.ESCALATED;
    case r'MERGED':
      return Enum$SupervisorState.MERGED;
    case r'REFUSED':
      return Enum$SupervisorState.REFUSED;
    case r'COMPLETED':
      return Enum$SupervisorState.COMPLETED;
    default:
      return Enum$SupervisorState.$unknown;
  }
}

enum Enum$SupervisorActionCode {
  CONTINUE,
  RETRY,
  BLOCK,
  ESCALATE,
  REQUEST_REWORK,
  MERGE,
  REFUSE,
  START_TASK,
  $unknown;

  factory Enum$SupervisorActionCode.fromJson(String value) =>
      fromJson$Enum$SupervisorActionCode(value);

  String toJson() => toJson$Enum$SupervisorActionCode(this);
}

String toJson$Enum$SupervisorActionCode(Enum$SupervisorActionCode e) {
  switch (e) {
    case Enum$SupervisorActionCode.CONTINUE:
      return r'CONTINUE';
    case Enum$SupervisorActionCode.RETRY:
      return r'RETRY';
    case Enum$SupervisorActionCode.BLOCK:
      return r'BLOCK';
    case Enum$SupervisorActionCode.ESCALATE:
      return r'ESCALATE';
    case Enum$SupervisorActionCode.REQUEST_REWORK:
      return r'REQUEST_REWORK';
    case Enum$SupervisorActionCode.MERGE:
      return r'MERGE';
    case Enum$SupervisorActionCode.REFUSE:
      return r'REFUSE';
    case Enum$SupervisorActionCode.START_TASK:
      return r'START_TASK';
    case Enum$SupervisorActionCode.$unknown:
      return r'$unknown';
  }
}

Enum$SupervisorActionCode fromJson$Enum$SupervisorActionCode(String value) {
  switch (value) {
    case r'CONTINUE':
      return Enum$SupervisorActionCode.CONTINUE;
    case r'RETRY':
      return Enum$SupervisorActionCode.RETRY;
    case r'BLOCK':
      return Enum$SupervisorActionCode.BLOCK;
    case r'ESCALATE':
      return Enum$SupervisorActionCode.ESCALATE;
    case r'REQUEST_REWORK':
      return Enum$SupervisorActionCode.REQUEST_REWORK;
    case r'MERGE':
      return Enum$SupervisorActionCode.MERGE;
    case r'REFUSE':
      return Enum$SupervisorActionCode.REFUSE;
    case r'START_TASK':
      return Enum$SupervisorActionCode.START_TASK;
    default:
      return Enum$SupervisorActionCode.$unknown;
  }
}

enum Enum$SupervisorReasonCode {
  JOB_ADMITTED,
  EXECUTION_PROGRESSED,
  EXECUTION_SUCCEEDED,
  EXECUTION_FAILED_RETRY,
  EXECUTION_FAILED_MAX_RETRIES,
  EXECUTION_FAILED_TERMINAL,
  TRACKER_ATTENTION_REQUIRED,
  TRACKER_ATTENTION_CLEARED,
  SCM_ATTENTION_REQUIRED,
  SCM_ATTENTION_CLEARED,
  PR_CONFLICT_DETECTED,
  PR_REVIEW_CHANGES_REQUESTED,
  PR_CHECKS_FAILED,
  PR_CHECKS_PASSED,
  PR_MERGE_APPROVED,
  PR_MERGE_REFUSED,
  ISSUE_AWAITING_APPROVAL,
  ISSUE_TASK_KICKOFF,
  MANUAL_OVERRIDE,
  POLICY_DEFAULT_CONTINUE,
  $unknown;

  factory Enum$SupervisorReasonCode.fromJson(String value) =>
      fromJson$Enum$SupervisorReasonCode(value);

  String toJson() => toJson$Enum$SupervisorReasonCode(this);
}

String toJson$Enum$SupervisorReasonCode(Enum$SupervisorReasonCode e) {
  switch (e) {
    case Enum$SupervisorReasonCode.JOB_ADMITTED:
      return r'JOB_ADMITTED';
    case Enum$SupervisorReasonCode.EXECUTION_PROGRESSED:
      return r'EXECUTION_PROGRESSED';
    case Enum$SupervisorReasonCode.EXECUTION_SUCCEEDED:
      return r'EXECUTION_SUCCEEDED';
    case Enum$SupervisorReasonCode.EXECUTION_FAILED_RETRY:
      return r'EXECUTION_FAILED_RETRY';
    case Enum$SupervisorReasonCode.EXECUTION_FAILED_MAX_RETRIES:
      return r'EXECUTION_FAILED_MAX_RETRIES';
    case Enum$SupervisorReasonCode.EXECUTION_FAILED_TERMINAL:
      return r'EXECUTION_FAILED_TERMINAL';
    case Enum$SupervisorReasonCode.TRACKER_ATTENTION_REQUIRED:
      return r'TRACKER_ATTENTION_REQUIRED';
    case Enum$SupervisorReasonCode.TRACKER_ATTENTION_CLEARED:
      return r'TRACKER_ATTENTION_CLEARED';
    case Enum$SupervisorReasonCode.SCM_ATTENTION_REQUIRED:
      return r'SCM_ATTENTION_REQUIRED';
    case Enum$SupervisorReasonCode.SCM_ATTENTION_CLEARED:
      return r'SCM_ATTENTION_CLEARED';
    case Enum$SupervisorReasonCode.PR_CONFLICT_DETECTED:
      return r'PR_CONFLICT_DETECTED';
    case Enum$SupervisorReasonCode.PR_REVIEW_CHANGES_REQUESTED:
      return r'PR_REVIEW_CHANGES_REQUESTED';
    case Enum$SupervisorReasonCode.PR_CHECKS_FAILED:
      return r'PR_CHECKS_FAILED';
    case Enum$SupervisorReasonCode.PR_CHECKS_PASSED:
      return r'PR_CHECKS_PASSED';
    case Enum$SupervisorReasonCode.PR_MERGE_APPROVED:
      return r'PR_MERGE_APPROVED';
    case Enum$SupervisorReasonCode.PR_MERGE_REFUSED:
      return r'PR_MERGE_REFUSED';
    case Enum$SupervisorReasonCode.ISSUE_AWAITING_APPROVAL:
      return r'ISSUE_AWAITING_APPROVAL';
    case Enum$SupervisorReasonCode.ISSUE_TASK_KICKOFF:
      return r'ISSUE_TASK_KICKOFF';
    case Enum$SupervisorReasonCode.MANUAL_OVERRIDE:
      return r'MANUAL_OVERRIDE';
    case Enum$SupervisorReasonCode.POLICY_DEFAULT_CONTINUE:
      return r'POLICY_DEFAULT_CONTINUE';
    case Enum$SupervisorReasonCode.$unknown:
      return r'$unknown';
  }
}

Enum$SupervisorReasonCode fromJson$Enum$SupervisorReasonCode(String value) {
  switch (value) {
    case r'JOB_ADMITTED':
      return Enum$SupervisorReasonCode.JOB_ADMITTED;
    case r'EXECUTION_PROGRESSED':
      return Enum$SupervisorReasonCode.EXECUTION_PROGRESSED;
    case r'EXECUTION_SUCCEEDED':
      return Enum$SupervisorReasonCode.EXECUTION_SUCCEEDED;
    case r'EXECUTION_FAILED_RETRY':
      return Enum$SupervisorReasonCode.EXECUTION_FAILED_RETRY;
    case r'EXECUTION_FAILED_MAX_RETRIES':
      return Enum$SupervisorReasonCode.EXECUTION_FAILED_MAX_RETRIES;
    case r'EXECUTION_FAILED_TERMINAL':
      return Enum$SupervisorReasonCode.EXECUTION_FAILED_TERMINAL;
    case r'TRACKER_ATTENTION_REQUIRED':
      return Enum$SupervisorReasonCode.TRACKER_ATTENTION_REQUIRED;
    case r'TRACKER_ATTENTION_CLEARED':
      return Enum$SupervisorReasonCode.TRACKER_ATTENTION_CLEARED;
    case r'SCM_ATTENTION_REQUIRED':
      return Enum$SupervisorReasonCode.SCM_ATTENTION_REQUIRED;
    case r'SCM_ATTENTION_CLEARED':
      return Enum$SupervisorReasonCode.SCM_ATTENTION_CLEARED;
    case r'PR_CONFLICT_DETECTED':
      return Enum$SupervisorReasonCode.PR_CONFLICT_DETECTED;
    case r'PR_REVIEW_CHANGES_REQUESTED':
      return Enum$SupervisorReasonCode.PR_REVIEW_CHANGES_REQUESTED;
    case r'PR_CHECKS_FAILED':
      return Enum$SupervisorReasonCode.PR_CHECKS_FAILED;
    case r'PR_CHECKS_PASSED':
      return Enum$SupervisorReasonCode.PR_CHECKS_PASSED;
    case r'PR_MERGE_APPROVED':
      return Enum$SupervisorReasonCode.PR_MERGE_APPROVED;
    case r'PR_MERGE_REFUSED':
      return Enum$SupervisorReasonCode.PR_MERGE_REFUSED;
    case r'ISSUE_AWAITING_APPROVAL':
      return Enum$SupervisorReasonCode.ISSUE_AWAITING_APPROVAL;
    case r'ISSUE_TASK_KICKOFF':
      return Enum$SupervisorReasonCode.ISSUE_TASK_KICKOFF;
    case r'MANUAL_OVERRIDE':
      return Enum$SupervisorReasonCode.MANUAL_OVERRIDE;
    case r'POLICY_DEFAULT_CONTINUE':
      return Enum$SupervisorReasonCode.POLICY_DEFAULT_CONTINUE;
    default:
      return Enum$SupervisorReasonCode.$unknown;
  }
}

enum Enum$SupervisorAttentionZone {
  NONE,
  TRACKER,
  SCM,
  EXECUTION,
  $unknown;

  factory Enum$SupervisorAttentionZone.fromJson(String value) =>
      fromJson$Enum$SupervisorAttentionZone(value);

  String toJson() => toJson$Enum$SupervisorAttentionZone(this);
}

String toJson$Enum$SupervisorAttentionZone(Enum$SupervisorAttentionZone e) {
  switch (e) {
    case Enum$SupervisorAttentionZone.NONE:
      return r'NONE';
    case Enum$SupervisorAttentionZone.TRACKER:
      return r'TRACKER';
    case Enum$SupervisorAttentionZone.SCM:
      return r'SCM';
    case Enum$SupervisorAttentionZone.EXECUTION:
      return r'EXECUTION';
    case Enum$SupervisorAttentionZone.$unknown:
      return r'$unknown';
  }
}

Enum$SupervisorAttentionZone fromJson$Enum$SupervisorAttentionZone(
  String value,
) {
  switch (value) {
    case r'NONE':
      return Enum$SupervisorAttentionZone.NONE;
    case r'TRACKER':
      return Enum$SupervisorAttentionZone.TRACKER;
    case r'SCM':
      return Enum$SupervisorAttentionZone.SCM;
    case r'EXECUTION':
      return Enum$SupervisorAttentionZone.EXECUTION;
    default:
      return Enum$SupervisorAttentionZone.$unknown;
  }
}

enum Enum$FailureClass {
  UNKNOWN,
  TRANSIENT,
  TERMINAL,
  $unknown;

  factory Enum$FailureClass.fromJson(String value) =>
      fromJson$Enum$FailureClass(value);

  String toJson() => toJson$Enum$FailureClass(this);
}

String toJson$Enum$FailureClass(Enum$FailureClass e) {
  switch (e) {
    case Enum$FailureClass.UNKNOWN:
      return r'UNKNOWN';
    case Enum$FailureClass.TRANSIENT:
      return r'TRANSIENT';
    case Enum$FailureClass.TERMINAL:
      return r'TERMINAL';
    case Enum$FailureClass.$unknown:
      return r'$unknown';
  }
}

Enum$FailureClass fromJson$Enum$FailureClass(String value) {
  switch (value) {
    case r'UNKNOWN':
      return Enum$FailureClass.UNKNOWN;
    case r'TRANSIENT':
      return Enum$FailureClass.TRANSIENT;
    case r'TERMINAL':
      return Enum$FailureClass.TERMINAL;
    default:
      return Enum$FailureClass.$unknown;
  }
}
