import 'scm.graphql.dart';

class Input$IngestionBoardSourceInput {
  factory Input$IngestionBoardSourceInput({
    required Enum$TrackerSourceKind kind,
    String? location,
    String? boardID,
  }) => Input$IngestionBoardSourceInput._({
    r'kind': kind,
    if (location != null) r'location': location,
    if (boardID != null) r'boardID': boardID,
  });

  Input$IngestionBoardSourceInput._(this._$data);

  factory Input$IngestionBoardSourceInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$kind = data['kind'];
    result$data['kind'] = fromJson$Enum$TrackerSourceKind((l$kind as String));
    if (data.containsKey('location')) {
      final l$location = data['location'];
      result$data['location'] = (l$location as String?);
    }
    if (data.containsKey('boardID')) {
      final l$boardID = data['boardID'];
      result$data['boardID'] = (l$boardID as String?);
    }
    return Input$IngestionBoardSourceInput._(result$data);
  }

  Map<String, dynamic> _$data;

  Enum$TrackerSourceKind get kind => (_$data['kind'] as Enum$TrackerSourceKind);

  String? get location => (_$data['location'] as String?);

  String? get boardID => (_$data['boardID'] as String?);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$kind = kind;
    result$data['kind'] = toJson$Enum$TrackerSourceKind(l$kind);
    if (_$data.containsKey('location')) {
      final l$location = location;
      result$data['location'] = l$location;
    }
    if (_$data.containsKey('boardID')) {
      final l$boardID = boardID;
      result$data['boardID'] = l$boardID;
    }
    return result$data;
  }

  CopyWith$Input$IngestionBoardSourceInput<Input$IngestionBoardSourceInput>
  get copyWith => CopyWith$Input$IngestionBoardSourceInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$IngestionBoardSourceInput ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$kind = kind;
    final lOther$kind = other.kind;
    if (l$kind != lOther$kind) {
      return false;
    }
    final l$location = location;
    final lOther$location = other.location;
    if (_$data.containsKey('location') !=
        other._$data.containsKey('location')) {
      return false;
    }
    if (l$location != lOther$location) {
      return false;
    }
    final l$boardID = boardID;
    final lOther$boardID = other.boardID;
    if (_$data.containsKey('boardID') != other._$data.containsKey('boardID')) {
      return false;
    }
    if (l$boardID != lOther$boardID) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$kind = kind;
    final l$location = location;
    final l$boardID = boardID;
    return Object.hashAll([
      l$kind,
      _$data.containsKey('location') ? l$location : const {},
      _$data.containsKey('boardID') ? l$boardID : const {},
    ]);
  }
}

abstract class CopyWith$Input$IngestionBoardSourceInput<TRes> {
  factory CopyWith$Input$IngestionBoardSourceInput(
    Input$IngestionBoardSourceInput instance,
    TRes Function(Input$IngestionBoardSourceInput) then,
  ) = _CopyWithImpl$Input$IngestionBoardSourceInput;

  factory CopyWith$Input$IngestionBoardSourceInput.stub(TRes res) =
      _CopyWithStubImpl$Input$IngestionBoardSourceInput;

  TRes call({Enum$TrackerSourceKind? kind, String? location, String? boardID});
}

class _CopyWithImpl$Input$IngestionBoardSourceInput<TRes>
    implements CopyWith$Input$IngestionBoardSourceInput<TRes> {
  _CopyWithImpl$Input$IngestionBoardSourceInput(this._instance, this._then);

  final Input$IngestionBoardSourceInput _instance;

  final TRes Function(Input$IngestionBoardSourceInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? kind = _undefined,
    Object? location = _undefined,
    Object? boardID = _undefined,
  }) => _then(
    Input$IngestionBoardSourceInput._({
      ..._instance._$data,
      if (kind != _undefined && kind != null)
        'kind': (kind as Enum$TrackerSourceKind),
      if (location != _undefined) 'location': (location as String?),
      if (boardID != _undefined) 'boardID': (boardID as String?),
    }),
  );
}

class _CopyWithStubImpl$Input$IngestionBoardSourceInput<TRes>
    implements CopyWith$Input$IngestionBoardSourceInput<TRes> {
  _CopyWithStubImpl$Input$IngestionBoardSourceInput(this._res);

  TRes _res;

  call({Enum$TrackerSourceKind? kind, String? location, String? boardID}) =>
      _res;
}

class Input$EnqueueIngestionWorkflowInput {
  factory Input$EnqueueIngestionWorkflowInput({
    required String runID,
    required String taskID,
    required String jobID,
    required String idempotencyKey,
    required String prompt,
    required String projectID,
    required String workflowID,
    required Input$IngestionBoardSourceInput boardSource,
  }) => Input$EnqueueIngestionWorkflowInput._({
    r'runID': runID,
    r'taskID': taskID,
    r'jobID': jobID,
    r'idempotencyKey': idempotencyKey,
    r'prompt': prompt,
    r'projectID': projectID,
    r'workflowID': workflowID,
    r'boardSource': boardSource,
  });

  Input$EnqueueIngestionWorkflowInput._(this._$data);

  factory Input$EnqueueIngestionWorkflowInput.fromJson(
    Map<String, dynamic> data,
  ) {
    final result$data = <String, dynamic>{};
    final l$runID = data['runID'];
    result$data['runID'] = (l$runID as String);
    final l$taskID = data['taskID'];
    result$data['taskID'] = (l$taskID as String);
    final l$jobID = data['jobID'];
    result$data['jobID'] = (l$jobID as String);
    final l$idempotencyKey = data['idempotencyKey'];
    result$data['idempotencyKey'] = (l$idempotencyKey as String);
    final l$prompt = data['prompt'];
    result$data['prompt'] = (l$prompt as String);
    final l$projectID = data['projectID'];
    result$data['projectID'] = (l$projectID as String);
    final l$workflowID = data['workflowID'];
    result$data['workflowID'] = (l$workflowID as String);
    final l$boardSource = data['boardSource'];
    result$data['boardSource'] = Input$IngestionBoardSourceInput.fromJson(
      (l$boardSource as Map<String, dynamic>),
    );
    return Input$EnqueueIngestionWorkflowInput._(result$data);
  }

  Map<String, dynamic> _$data;

  String get runID => (_$data['runID'] as String);

  String get taskID => (_$data['taskID'] as String);

  String get jobID => (_$data['jobID'] as String);

  String get idempotencyKey => (_$data['idempotencyKey'] as String);

  String get prompt => (_$data['prompt'] as String);

  String get projectID => (_$data['projectID'] as String);

  String get workflowID => (_$data['workflowID'] as String);

  Input$IngestionBoardSourceInput get boardSource =>
      (_$data['boardSource'] as Input$IngestionBoardSourceInput);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$runID = runID;
    result$data['runID'] = l$runID;
    final l$taskID = taskID;
    result$data['taskID'] = l$taskID;
    final l$jobID = jobID;
    result$data['jobID'] = l$jobID;
    final l$idempotencyKey = idempotencyKey;
    result$data['idempotencyKey'] = l$idempotencyKey;
    final l$prompt = prompt;
    result$data['prompt'] = l$prompt;
    final l$projectID = projectID;
    result$data['projectID'] = l$projectID;
    final l$workflowID = workflowID;
    result$data['workflowID'] = l$workflowID;
    final l$boardSource = boardSource;
    result$data['boardSource'] = l$boardSource.toJson();
    return result$data;
  }

  CopyWith$Input$EnqueueIngestionWorkflowInput<
    Input$EnqueueIngestionWorkflowInput
  >
  get copyWith => CopyWith$Input$EnqueueIngestionWorkflowInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$EnqueueIngestionWorkflowInput ||
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
    final l$idempotencyKey = idempotencyKey;
    final lOther$idempotencyKey = other.idempotencyKey;
    if (l$idempotencyKey != lOther$idempotencyKey) {
      return false;
    }
    final l$prompt = prompt;
    final lOther$prompt = other.prompt;
    if (l$prompt != lOther$prompt) {
      return false;
    }
    final l$projectID = projectID;
    final lOther$projectID = other.projectID;
    if (l$projectID != lOther$projectID) {
      return false;
    }
    final l$workflowID = workflowID;
    final lOther$workflowID = other.workflowID;
    if (l$workflowID != lOther$workflowID) {
      return false;
    }
    final l$boardSource = boardSource;
    final lOther$boardSource = other.boardSource;
    if (l$boardSource != lOther$boardSource) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$runID = runID;
    final l$taskID = taskID;
    final l$jobID = jobID;
    final l$idempotencyKey = idempotencyKey;
    final l$prompt = prompt;
    final l$projectID = projectID;
    final l$workflowID = workflowID;
    final l$boardSource = boardSource;
    return Object.hashAll([
      l$runID,
      l$taskID,
      l$jobID,
      l$idempotencyKey,
      l$prompt,
      l$projectID,
      l$workflowID,
      l$boardSource,
    ]);
  }
}

abstract class CopyWith$Input$EnqueueIngestionWorkflowInput<TRes> {
  factory CopyWith$Input$EnqueueIngestionWorkflowInput(
    Input$EnqueueIngestionWorkflowInput instance,
    TRes Function(Input$EnqueueIngestionWorkflowInput) then,
  ) = _CopyWithImpl$Input$EnqueueIngestionWorkflowInput;

  factory CopyWith$Input$EnqueueIngestionWorkflowInput.stub(TRes res) =
      _CopyWithStubImpl$Input$EnqueueIngestionWorkflowInput;

  TRes call({
    String? runID,
    String? taskID,
    String? jobID,
    String? idempotencyKey,
    String? prompt,
    String? projectID,
    String? workflowID,
    Input$IngestionBoardSourceInput? boardSource,
  });
  CopyWith$Input$IngestionBoardSourceInput<TRes> get boardSource;
}

class _CopyWithImpl$Input$EnqueueIngestionWorkflowInput<TRes>
    implements CopyWith$Input$EnqueueIngestionWorkflowInput<TRes> {
  _CopyWithImpl$Input$EnqueueIngestionWorkflowInput(this._instance, this._then);

  final Input$EnqueueIngestionWorkflowInput _instance;

  final TRes Function(Input$EnqueueIngestionWorkflowInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? runID = _undefined,
    Object? taskID = _undefined,
    Object? jobID = _undefined,
    Object? idempotencyKey = _undefined,
    Object? prompt = _undefined,
    Object? projectID = _undefined,
    Object? workflowID = _undefined,
    Object? boardSource = _undefined,
  }) => _then(
    Input$EnqueueIngestionWorkflowInput._({
      ..._instance._$data,
      if (runID != _undefined && runID != null) 'runID': (runID as String),
      if (taskID != _undefined && taskID != null) 'taskID': (taskID as String),
      if (jobID != _undefined && jobID != null) 'jobID': (jobID as String),
      if (idempotencyKey != _undefined && idempotencyKey != null)
        'idempotencyKey': (idempotencyKey as String),
      if (prompt != _undefined && prompt != null) 'prompt': (prompt as String),
      if (projectID != _undefined && projectID != null)
        'projectID': (projectID as String),
      if (workflowID != _undefined && workflowID != null)
        'workflowID': (workflowID as String),
      if (boardSource != _undefined && boardSource != null)
        'boardSource': (boardSource as Input$IngestionBoardSourceInput),
    }),
  );

  CopyWith$Input$IngestionBoardSourceInput<TRes> get boardSource {
    final local$boardSource = _instance.boardSource;
    return CopyWith$Input$IngestionBoardSourceInput(
      local$boardSource,
      (e) => call(boardSource: e),
    );
  }
}

class _CopyWithStubImpl$Input$EnqueueIngestionWorkflowInput<TRes>
    implements CopyWith$Input$EnqueueIngestionWorkflowInput<TRes> {
  _CopyWithStubImpl$Input$EnqueueIngestionWorkflowInput(this._res);

  TRes _res;

  call({
    String? runID,
    String? taskID,
    String? jobID,
    String? idempotencyKey,
    String? prompt,
    String? projectID,
    String? workflowID,
    Input$IngestionBoardSourceInput? boardSource,
  }) => _res;

  CopyWith$Input$IngestionBoardSourceInput<TRes> get boardSource =>
      CopyWith$Input$IngestionBoardSourceInput.stub(_res);
}

class Input$ApproveIssueIntakeInput {
  factory Input$ApproveIssueIntakeInput({
    required String runID,
    required String taskID,
    required String jobID,
    required String source,
    required String issueReference,
    required String approvedBy,
  }) => Input$ApproveIssueIntakeInput._({
    r'runID': runID,
    r'taskID': taskID,
    r'jobID': jobID,
    r'source': source,
    r'issueReference': issueReference,
    r'approvedBy': approvedBy,
  });

  Input$ApproveIssueIntakeInput._(this._$data);

  factory Input$ApproveIssueIntakeInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$runID = data['runID'];
    result$data['runID'] = (l$runID as String);
    final l$taskID = data['taskID'];
    result$data['taskID'] = (l$taskID as String);
    final l$jobID = data['jobID'];
    result$data['jobID'] = (l$jobID as String);
    final l$source = data['source'];
    result$data['source'] = (l$source as String);
    final l$issueReference = data['issueReference'];
    result$data['issueReference'] = (l$issueReference as String);
    final l$approvedBy = data['approvedBy'];
    result$data['approvedBy'] = (l$approvedBy as String);
    return Input$ApproveIssueIntakeInput._(result$data);
  }

  Map<String, dynamic> _$data;

  String get runID => (_$data['runID'] as String);

  String get taskID => (_$data['taskID'] as String);

  String get jobID => (_$data['jobID'] as String);

  String get source => (_$data['source'] as String);

  String get issueReference => (_$data['issueReference'] as String);

  String get approvedBy => (_$data['approvedBy'] as String);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$runID = runID;
    result$data['runID'] = l$runID;
    final l$taskID = taskID;
    result$data['taskID'] = l$taskID;
    final l$jobID = jobID;
    result$data['jobID'] = l$jobID;
    final l$source = source;
    result$data['source'] = l$source;
    final l$issueReference = issueReference;
    result$data['issueReference'] = l$issueReference;
    final l$approvedBy = approvedBy;
    result$data['approvedBy'] = l$approvedBy;
    return result$data;
  }

  CopyWith$Input$ApproveIssueIntakeInput<Input$ApproveIssueIntakeInput>
  get copyWith => CopyWith$Input$ApproveIssueIntakeInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$ApproveIssueIntakeInput ||
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
    final l$source = source;
    final lOther$source = other.source;
    if (l$source != lOther$source) {
      return false;
    }
    final l$issueReference = issueReference;
    final lOther$issueReference = other.issueReference;
    if (l$issueReference != lOther$issueReference) {
      return false;
    }
    final l$approvedBy = approvedBy;
    final lOther$approvedBy = other.approvedBy;
    if (l$approvedBy != lOther$approvedBy) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$runID = runID;
    final l$taskID = taskID;
    final l$jobID = jobID;
    final l$source = source;
    final l$issueReference = issueReference;
    final l$approvedBy = approvedBy;
    return Object.hashAll([
      l$runID,
      l$taskID,
      l$jobID,
      l$source,
      l$issueReference,
      l$approvedBy,
    ]);
  }
}

abstract class CopyWith$Input$ApproveIssueIntakeInput<TRes> {
  factory CopyWith$Input$ApproveIssueIntakeInput(
    Input$ApproveIssueIntakeInput instance,
    TRes Function(Input$ApproveIssueIntakeInput) then,
  ) = _CopyWithImpl$Input$ApproveIssueIntakeInput;

  factory CopyWith$Input$ApproveIssueIntakeInput.stub(TRes res) =
      _CopyWithStubImpl$Input$ApproveIssueIntakeInput;

  TRes call({
    String? runID,
    String? taskID,
    String? jobID,
    String? source,
    String? issueReference,
    String? approvedBy,
  });
}

class _CopyWithImpl$Input$ApproveIssueIntakeInput<TRes>
    implements CopyWith$Input$ApproveIssueIntakeInput<TRes> {
  _CopyWithImpl$Input$ApproveIssueIntakeInput(this._instance, this._then);

  final Input$ApproveIssueIntakeInput _instance;

  final TRes Function(Input$ApproveIssueIntakeInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? runID = _undefined,
    Object? taskID = _undefined,
    Object? jobID = _undefined,
    Object? source = _undefined,
    Object? issueReference = _undefined,
    Object? approvedBy = _undefined,
  }) => _then(
    Input$ApproveIssueIntakeInput._({
      ..._instance._$data,
      if (runID != _undefined && runID != null) 'runID': (runID as String),
      if (taskID != _undefined && taskID != null) 'taskID': (taskID as String),
      if (jobID != _undefined && jobID != null) 'jobID': (jobID as String),
      if (source != _undefined && source != null) 'source': (source as String),
      if (issueReference != _undefined && issueReference != null)
        'issueReference': (issueReference as String),
      if (approvedBy != _undefined && approvedBy != null)
        'approvedBy': (approvedBy as String),
    }),
  );
}

class _CopyWithStubImpl$Input$ApproveIssueIntakeInput<TRes>
    implements CopyWith$Input$ApproveIssueIntakeInput<TRes> {
  _CopyWithStubImpl$Input$ApproveIssueIntakeInput(this._res);

  TRes _res;

  call({
    String? runID,
    String? taskID,
    String? jobID,
    String? source,
    String? issueReference,
    String? approvedBy,
  }) => _res;
}

class Input$RequeueDeadLetterInput {
  factory Input$RequeueDeadLetterInput({
    required String queue,
    required String taskID,
  }) => Input$RequeueDeadLetterInput._({r'queue': queue, r'taskID': taskID});

  Input$RequeueDeadLetterInput._(this._$data);

  factory Input$RequeueDeadLetterInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$queue = data['queue'];
    result$data['queue'] = (l$queue as String);
    final l$taskID = data['taskID'];
    result$data['taskID'] = (l$taskID as String);
    return Input$RequeueDeadLetterInput._(result$data);
  }

  Map<String, dynamic> _$data;

  String get queue => (_$data['queue'] as String);

  String get taskID => (_$data['taskID'] as String);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$queue = queue;
    result$data['queue'] = l$queue;
    final l$taskID = taskID;
    result$data['taskID'] = l$taskID;
    return result$data;
  }

  CopyWith$Input$RequeueDeadLetterInput<Input$RequeueDeadLetterInput>
  get copyWith => CopyWith$Input$RequeueDeadLetterInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$RequeueDeadLetterInput ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$queue = queue;
    final lOther$queue = other.queue;
    if (l$queue != lOther$queue) {
      return false;
    }
    final l$taskID = taskID;
    final lOther$taskID = other.taskID;
    if (l$taskID != lOther$taskID) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$queue = queue;
    final l$taskID = taskID;
    return Object.hashAll([l$queue, l$taskID]);
  }
}

abstract class CopyWith$Input$RequeueDeadLetterInput<TRes> {
  factory CopyWith$Input$RequeueDeadLetterInput(
    Input$RequeueDeadLetterInput instance,
    TRes Function(Input$RequeueDeadLetterInput) then,
  ) = _CopyWithImpl$Input$RequeueDeadLetterInput;

  factory CopyWith$Input$RequeueDeadLetterInput.stub(TRes res) =
      _CopyWithStubImpl$Input$RequeueDeadLetterInput;

  TRes call({String? queue, String? taskID});
}

class _CopyWithImpl$Input$RequeueDeadLetterInput<TRes>
    implements CopyWith$Input$RequeueDeadLetterInput<TRes> {
  _CopyWithImpl$Input$RequeueDeadLetterInput(this._instance, this._then);

  final Input$RequeueDeadLetterInput _instance;

  final TRes Function(Input$RequeueDeadLetterInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({Object? queue = _undefined, Object? taskID = _undefined}) => _then(
    Input$RequeueDeadLetterInput._({
      ..._instance._$data,
      if (queue != _undefined && queue != null) 'queue': (queue as String),
      if (taskID != _undefined && taskID != null) 'taskID': (taskID as String),
    }),
  );
}

class _CopyWithStubImpl$Input$RequeueDeadLetterInput<TRes>
    implements CopyWith$Input$RequeueDeadLetterInput<TRes> {
  _CopyWithStubImpl$Input$RequeueDeadLetterInput(this._res);

  TRes _res;

  call({String? queue, String? taskID}) => _res;
}

class Input$UpsertProjectSetupInput {
  factory Input$UpsertProjectSetupInput({
    required String projectID,
    required String projectName,
    required Enum$SCMProvider scmProvider,
    required String repositoryURL,
    required Enum$TrackerSourceKind trackerProvider,
    String? trackerLocation,
    String? trackerBoardID,
  }) => Input$UpsertProjectSetupInput._({
    r'projectID': projectID,
    r'projectName': projectName,
    r'scmProvider': scmProvider,
    r'repositoryURL': repositoryURL,
    r'trackerProvider': trackerProvider,
    if (trackerLocation != null) r'trackerLocation': trackerLocation,
    if (trackerBoardID != null) r'trackerBoardID': trackerBoardID,
  });

  Input$UpsertProjectSetupInput._(this._$data);

  factory Input$UpsertProjectSetupInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$projectID = data['projectID'];
    result$data['projectID'] = (l$projectID as String);
    final l$projectName = data['projectName'];
    result$data['projectName'] = (l$projectName as String);
    final l$scmProvider = data['scmProvider'];
    result$data['scmProvider'] = fromJson$Enum$SCMProvider(
      (l$scmProvider as String),
    );
    final l$repositoryURL = data['repositoryURL'];
    result$data['repositoryURL'] = (l$repositoryURL as String);
    final l$trackerProvider = data['trackerProvider'];
    result$data['trackerProvider'] = fromJson$Enum$TrackerSourceKind(
      (l$trackerProvider as String),
    );
    if (data.containsKey('trackerLocation')) {
      final l$trackerLocation = data['trackerLocation'];
      result$data['trackerLocation'] = (l$trackerLocation as String?);
    }
    if (data.containsKey('trackerBoardID')) {
      final l$trackerBoardID = data['trackerBoardID'];
      result$data['trackerBoardID'] = (l$trackerBoardID as String?);
    }
    return Input$UpsertProjectSetupInput._(result$data);
  }

  Map<String, dynamic> _$data;

  String get projectID => (_$data['projectID'] as String);

  String get projectName => (_$data['projectName'] as String);

  Enum$SCMProvider get scmProvider =>
      (_$data['scmProvider'] as Enum$SCMProvider);

  String get repositoryURL => (_$data['repositoryURL'] as String);

  Enum$TrackerSourceKind get trackerProvider =>
      (_$data['trackerProvider'] as Enum$TrackerSourceKind);

  String? get trackerLocation => (_$data['trackerLocation'] as String?);

  String? get trackerBoardID => (_$data['trackerBoardID'] as String?);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$projectID = projectID;
    result$data['projectID'] = l$projectID;
    final l$projectName = projectName;
    result$data['projectName'] = l$projectName;
    final l$scmProvider = scmProvider;
    result$data['scmProvider'] = toJson$Enum$SCMProvider(l$scmProvider);
    final l$repositoryURL = repositoryURL;
    result$data['repositoryURL'] = l$repositoryURL;
    final l$trackerProvider = trackerProvider;
    result$data['trackerProvider'] = toJson$Enum$TrackerSourceKind(
      l$trackerProvider,
    );
    if (_$data.containsKey('trackerLocation')) {
      final l$trackerLocation = trackerLocation;
      result$data['trackerLocation'] = l$trackerLocation;
    }
    if (_$data.containsKey('trackerBoardID')) {
      final l$trackerBoardID = trackerBoardID;
      result$data['trackerBoardID'] = l$trackerBoardID;
    }
    return result$data;
  }

  CopyWith$Input$UpsertProjectSetupInput<Input$UpsertProjectSetupInput>
  get copyWith => CopyWith$Input$UpsertProjectSetupInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$UpsertProjectSetupInput ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$projectID = projectID;
    final lOther$projectID = other.projectID;
    if (l$projectID != lOther$projectID) {
      return false;
    }
    final l$projectName = projectName;
    final lOther$projectName = other.projectName;
    if (l$projectName != lOther$projectName) {
      return false;
    }
    final l$scmProvider = scmProvider;
    final lOther$scmProvider = other.scmProvider;
    if (l$scmProvider != lOther$scmProvider) {
      return false;
    }
    final l$repositoryURL = repositoryURL;
    final lOther$repositoryURL = other.repositoryURL;
    if (l$repositoryURL != lOther$repositoryURL) {
      return false;
    }
    final l$trackerProvider = trackerProvider;
    final lOther$trackerProvider = other.trackerProvider;
    if (l$trackerProvider != lOther$trackerProvider) {
      return false;
    }
    final l$trackerLocation = trackerLocation;
    final lOther$trackerLocation = other.trackerLocation;
    if (_$data.containsKey('trackerLocation') !=
        other._$data.containsKey('trackerLocation')) {
      return false;
    }
    if (l$trackerLocation != lOther$trackerLocation) {
      return false;
    }
    final l$trackerBoardID = trackerBoardID;
    final lOther$trackerBoardID = other.trackerBoardID;
    if (_$data.containsKey('trackerBoardID') !=
        other._$data.containsKey('trackerBoardID')) {
      return false;
    }
    if (l$trackerBoardID != lOther$trackerBoardID) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$projectID = projectID;
    final l$projectName = projectName;
    final l$scmProvider = scmProvider;
    final l$repositoryURL = repositoryURL;
    final l$trackerProvider = trackerProvider;
    final l$trackerLocation = trackerLocation;
    final l$trackerBoardID = trackerBoardID;
    return Object.hashAll([
      l$projectID,
      l$projectName,
      l$scmProvider,
      l$repositoryURL,
      l$trackerProvider,
      _$data.containsKey('trackerLocation') ? l$trackerLocation : const {},
      _$data.containsKey('trackerBoardID') ? l$trackerBoardID : const {},
    ]);
  }
}

abstract class CopyWith$Input$UpsertProjectSetupInput<TRes> {
  factory CopyWith$Input$UpsertProjectSetupInput(
    Input$UpsertProjectSetupInput instance,
    TRes Function(Input$UpsertProjectSetupInput) then,
  ) = _CopyWithImpl$Input$UpsertProjectSetupInput;

  factory CopyWith$Input$UpsertProjectSetupInput.stub(TRes res) =
      _CopyWithStubImpl$Input$UpsertProjectSetupInput;

  TRes call({
    String? projectID,
    String? projectName,
    Enum$SCMProvider? scmProvider,
    String? repositoryURL,
    Enum$TrackerSourceKind? trackerProvider,
    String? trackerLocation,
    String? trackerBoardID,
  });
}

class _CopyWithImpl$Input$UpsertProjectSetupInput<TRes>
    implements CopyWith$Input$UpsertProjectSetupInput<TRes> {
  _CopyWithImpl$Input$UpsertProjectSetupInput(this._instance, this._then);

  final Input$UpsertProjectSetupInput _instance;

  final TRes Function(Input$UpsertProjectSetupInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? projectID = _undefined,
    Object? projectName = _undefined,
    Object? scmProvider = _undefined,
    Object? repositoryURL = _undefined,
    Object? trackerProvider = _undefined,
    Object? trackerLocation = _undefined,
    Object? trackerBoardID = _undefined,
  }) => _then(
    Input$UpsertProjectSetupInput._({
      ..._instance._$data,
      if (projectID != _undefined && projectID != null)
        'projectID': (projectID as String),
      if (projectName != _undefined && projectName != null)
        'projectName': (projectName as String),
      if (scmProvider != _undefined && scmProvider != null)
        'scmProvider': (scmProvider as Enum$SCMProvider),
      if (repositoryURL != _undefined && repositoryURL != null)
        'repositoryURL': (repositoryURL as String),
      if (trackerProvider != _undefined && trackerProvider != null)
        'trackerProvider': (trackerProvider as Enum$TrackerSourceKind),
      if (trackerLocation != _undefined)
        'trackerLocation': (trackerLocation as String?),
      if (trackerBoardID != _undefined)
        'trackerBoardID': (trackerBoardID as String?),
    }),
  );
}

class _CopyWithStubImpl$Input$UpsertProjectSetupInput<TRes>
    implements CopyWith$Input$UpsertProjectSetupInput<TRes> {
  _CopyWithStubImpl$Input$UpsertProjectSetupInput(this._res);

  TRes _res;

  call({
    String? projectID,
    String? projectName,
    Enum$SCMProvider? scmProvider,
    String? repositoryURL,
    Enum$TrackerSourceKind? trackerProvider,
    String? trackerLocation,
    String? trackerBoardID,
  }) => _res;
}

enum Enum$JobKind {
  INGESTION_AGENT_RUN,
  AGENT_WORKFLOW_RUN,
  SCM_WORKFLOW_RUN,
  $unknown;

  factory Enum$JobKind.fromJson(String value) => fromJson$Enum$JobKind(value);

  String toJson() => toJson$Enum$JobKind(this);
}

String toJson$Enum$JobKind(Enum$JobKind e) {
  switch (e) {
    case Enum$JobKind.INGESTION_AGENT_RUN:
      return r'INGESTION_AGENT_RUN';
    case Enum$JobKind.AGENT_WORKFLOW_RUN:
      return r'AGENT_WORKFLOW_RUN';
    case Enum$JobKind.SCM_WORKFLOW_RUN:
      return r'SCM_WORKFLOW_RUN';
    case Enum$JobKind.$unknown:
      return r'$unknown';
  }
}

Enum$JobKind fromJson$Enum$JobKind(String value) {
  switch (value) {
    case r'INGESTION_AGENT_RUN':
      return Enum$JobKind.INGESTION_AGENT_RUN;
    case r'AGENT_WORKFLOW_RUN':
      return Enum$JobKind.AGENT_WORKFLOW_RUN;
    case r'SCM_WORKFLOW_RUN':
      return Enum$JobKind.SCM_WORKFLOW_RUN;
    default:
      return Enum$JobKind.$unknown;
  }
}

enum Enum$TrackerSourceKind {
  LOCAL_JSON,
  GITHUB_ISSUES,
  JIRA,
  LINEAR,
  $unknown;

  factory Enum$TrackerSourceKind.fromJson(String value) =>
      fromJson$Enum$TrackerSourceKind(value);

  String toJson() => toJson$Enum$TrackerSourceKind(this);
}

String toJson$Enum$TrackerSourceKind(Enum$TrackerSourceKind e) {
  switch (e) {
    case Enum$TrackerSourceKind.LOCAL_JSON:
      return r'LOCAL_JSON';
    case Enum$TrackerSourceKind.GITHUB_ISSUES:
      return r'GITHUB_ISSUES';
    case Enum$TrackerSourceKind.JIRA:
      return r'JIRA';
    case Enum$TrackerSourceKind.LINEAR:
      return r'LINEAR';
    case Enum$TrackerSourceKind.$unknown:
      return r'$unknown';
  }
}

Enum$TrackerSourceKind fromJson$Enum$TrackerSourceKind(String value) {
  switch (value) {
    case r'LOCAL_JSON':
      return Enum$TrackerSourceKind.LOCAL_JSON;
    case r'GITHUB_ISSUES':
      return Enum$TrackerSourceKind.GITHUB_ISSUES;
    case r'JIRA':
      return Enum$TrackerSourceKind.JIRA;
    case r'LINEAR':
      return Enum$TrackerSourceKind.LINEAR;
    default:
      return Enum$TrackerSourceKind.$unknown;
  }
}

enum Enum$StreamEventSource {
  ACP,
  SESSION_FILE,
  WORKER,
  $unknown;

  factory Enum$StreamEventSource.fromJson(String value) =>
      fromJson$Enum$StreamEventSource(value);

  String toJson() => toJson$Enum$StreamEventSource(this);
}

String toJson$Enum$StreamEventSource(Enum$StreamEventSource e) {
  switch (e) {
    case Enum$StreamEventSource.ACP:
      return r'ACP';
    case Enum$StreamEventSource.SESSION_FILE:
      return r'SESSION_FILE';
    case Enum$StreamEventSource.WORKER:
      return r'WORKER';
    case Enum$StreamEventSource.$unknown:
      return r'$unknown';
  }
}

Enum$StreamEventSource fromJson$Enum$StreamEventSource(String value) {
  switch (value) {
    case r'ACP':
      return Enum$StreamEventSource.ACP;
    case r'SESSION_FILE':
      return Enum$StreamEventSource.SESSION_FILE;
    case r'WORKER':
      return Enum$StreamEventSource.WORKER;
    default:
      return Enum$StreamEventSource.$unknown;
  }
}
