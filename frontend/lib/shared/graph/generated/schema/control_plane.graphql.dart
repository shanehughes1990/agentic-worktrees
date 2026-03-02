import 'scm.graphql.dart';

class Input$IngestionBoardSourceInput {
  factory Input$IngestionBoardSourceInput({
    required String boardID,
    required Enum$TrackerSourceKind kind,
    String? location,
    required bool appliesToAllRepositories,
    List<String>? repositoryIDs,
  }) => Input$IngestionBoardSourceInput._({
    r'boardID': boardID,
    r'kind': kind,
    if (location != null) r'location': location,
    r'appliesToAllRepositories': appliesToAllRepositories,
    if (repositoryIDs != null) r'repositoryIDs': repositoryIDs,
  });

  Input$IngestionBoardSourceInput._(this._$data);

  factory Input$IngestionBoardSourceInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$boardID = data['boardID'];
    result$data['boardID'] = (l$boardID as String);
    final l$kind = data['kind'];
    result$data['kind'] = fromJson$Enum$TrackerSourceKind((l$kind as String));
    if (data.containsKey('location')) {
      final l$location = data['location'];
      result$data['location'] = (l$location as String?);
    }
    final l$appliesToAllRepositories = data['appliesToAllRepositories'];
    result$data['appliesToAllRepositories'] =
        (l$appliesToAllRepositories as bool);
    if (data.containsKey('repositoryIDs')) {
      final l$repositoryIDs = data['repositoryIDs'];
      result$data['repositoryIDs'] = (l$repositoryIDs as List<dynamic>?)
          ?.map((e) => (e as String))
          .toList();
    }
    return Input$IngestionBoardSourceInput._(result$data);
  }

  Map<String, dynamic> _$data;

  String get boardID => (_$data['boardID'] as String);

  Enum$TrackerSourceKind get kind => (_$data['kind'] as Enum$TrackerSourceKind);

  String? get location => (_$data['location'] as String?);

  bool get appliesToAllRepositories =>
      (_$data['appliesToAllRepositories'] as bool);

  List<String>? get repositoryIDs => (_$data['repositoryIDs'] as List<String>?);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$boardID = boardID;
    result$data['boardID'] = l$boardID;
    final l$kind = kind;
    result$data['kind'] = toJson$Enum$TrackerSourceKind(l$kind);
    if (_$data.containsKey('location')) {
      final l$location = location;
      result$data['location'] = l$location;
    }
    final l$appliesToAllRepositories = appliesToAllRepositories;
    result$data['appliesToAllRepositories'] = l$appliesToAllRepositories;
    if (_$data.containsKey('repositoryIDs')) {
      final l$repositoryIDs = repositoryIDs;
      result$data['repositoryIDs'] = l$repositoryIDs?.map((e) => e).toList();
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
    final l$boardID = boardID;
    final lOther$boardID = other.boardID;
    if (l$boardID != lOther$boardID) {
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
    final l$appliesToAllRepositories = appliesToAllRepositories;
    final lOther$appliesToAllRepositories = other.appliesToAllRepositories;
    if (l$appliesToAllRepositories != lOther$appliesToAllRepositories) {
      return false;
    }
    final l$repositoryIDs = repositoryIDs;
    final lOther$repositoryIDs = other.repositoryIDs;
    if (_$data.containsKey('repositoryIDs') !=
        other._$data.containsKey('repositoryIDs')) {
      return false;
    }
    if (l$repositoryIDs != null && lOther$repositoryIDs != null) {
      if (l$repositoryIDs.length != lOther$repositoryIDs.length) {
        return false;
      }
      for (int i = 0; i < l$repositoryIDs.length; i++) {
        final l$repositoryIDs$entry = l$repositoryIDs[i];
        final lOther$repositoryIDs$entry = lOther$repositoryIDs[i];
        if (l$repositoryIDs$entry != lOther$repositoryIDs$entry) {
          return false;
        }
      }
    } else if (l$repositoryIDs != lOther$repositoryIDs) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$boardID = boardID;
    final l$kind = kind;
    final l$location = location;
    final l$appliesToAllRepositories = appliesToAllRepositories;
    final l$repositoryIDs = repositoryIDs;
    return Object.hashAll([
      l$boardID,
      l$kind,
      _$data.containsKey('location') ? l$location : const {},
      l$appliesToAllRepositories,
      _$data.containsKey('repositoryIDs')
          ? l$repositoryIDs == null
                ? null
                : Object.hashAll(l$repositoryIDs.map((v) => v))
          : const {},
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

  TRes call({
    String? boardID,
    Enum$TrackerSourceKind? kind,
    String? location,
    bool? appliesToAllRepositories,
    List<String>? repositoryIDs,
  });
}

class _CopyWithImpl$Input$IngestionBoardSourceInput<TRes>
    implements CopyWith$Input$IngestionBoardSourceInput<TRes> {
  _CopyWithImpl$Input$IngestionBoardSourceInput(this._instance, this._then);

  final Input$IngestionBoardSourceInput _instance;

  final TRes Function(Input$IngestionBoardSourceInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? boardID = _undefined,
    Object? kind = _undefined,
    Object? location = _undefined,
    Object? appliesToAllRepositories = _undefined,
    Object? repositoryIDs = _undefined,
  }) => _then(
    Input$IngestionBoardSourceInput._({
      ..._instance._$data,
      if (boardID != _undefined && boardID != null)
        'boardID': (boardID as String),
      if (kind != _undefined && kind != null)
        'kind': (kind as Enum$TrackerSourceKind),
      if (location != _undefined) 'location': (location as String?),
      if (appliesToAllRepositories != _undefined &&
          appliesToAllRepositories != null)
        'appliesToAllRepositories': (appliesToAllRepositories as bool),
      if (repositoryIDs != _undefined)
        'repositoryIDs': (repositoryIDs as List<String>?),
    }),
  );
}

class _CopyWithStubImpl$Input$IngestionBoardSourceInput<TRes>
    implements CopyWith$Input$IngestionBoardSourceInput<TRes> {
  _CopyWithStubImpl$Input$IngestionBoardSourceInput(this._res);

  TRes _res;

  call({
    String? boardID,
    Enum$TrackerSourceKind? kind,
    String? location,
    bool? appliesToAllRepositories,
    List<String>? repositoryIDs,
  }) => _res;
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
    required List<Input$IngestionBoardSourceInput> boardSources,
  }) => Input$EnqueueIngestionWorkflowInput._({
    r'runID': runID,
    r'taskID': taskID,
    r'jobID': jobID,
    r'idempotencyKey': idempotencyKey,
    r'prompt': prompt,
    r'projectID': projectID,
    r'workflowID': workflowID,
    r'boardSources': boardSources,
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
    final l$boardSources = data['boardSources'];
    result$data['boardSources'] = (l$boardSources as List<dynamic>)
        .map(
          (e) => Input$IngestionBoardSourceInput.fromJson(
            (e as Map<String, dynamic>),
          ),
        )
        .toList();
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

  List<Input$IngestionBoardSourceInput> get boardSources =>
      (_$data['boardSources'] as List<Input$IngestionBoardSourceInput>);

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
    final l$boardSources = boardSources;
    result$data['boardSources'] = l$boardSources
        .map((e) => e.toJson())
        .toList();
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
    final l$boardSources = boardSources;
    final lOther$boardSources = other.boardSources;
    if (l$boardSources.length != lOther$boardSources.length) {
      return false;
    }
    for (int i = 0; i < l$boardSources.length; i++) {
      final l$boardSources$entry = l$boardSources[i];
      final lOther$boardSources$entry = lOther$boardSources[i];
      if (l$boardSources$entry != lOther$boardSources$entry) {
        return false;
      }
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
    final l$boardSources = boardSources;
    return Object.hashAll([
      l$runID,
      l$taskID,
      l$jobID,
      l$idempotencyKey,
      l$prompt,
      l$projectID,
      l$workflowID,
      Object.hashAll(l$boardSources.map((v) => v)),
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
    List<Input$IngestionBoardSourceInput>? boardSources,
  });
  TRes boardSources(
    Iterable<Input$IngestionBoardSourceInput> Function(
      Iterable<
        CopyWith$Input$IngestionBoardSourceInput<
          Input$IngestionBoardSourceInput
        >
      >,
    )
    _fn,
  );
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
    Object? boardSources = _undefined,
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
      if (boardSources != _undefined && boardSources != null)
        'boardSources': (boardSources as List<Input$IngestionBoardSourceInput>),
    }),
  );

  TRes boardSources(
    Iterable<Input$IngestionBoardSourceInput> Function(
      Iterable<
        CopyWith$Input$IngestionBoardSourceInput<
          Input$IngestionBoardSourceInput
        >
      >,
    )
    _fn,
  ) => call(
    boardSources: _fn(
      _instance.boardSources.map(
        (e) => CopyWith$Input$IngestionBoardSourceInput(e, (i) => i),
      ),
    ).toList(),
  );
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
    List<Input$IngestionBoardSourceInput>? boardSources,
  }) => _res;

  boardSources(_fn) => _res;
}

class Input$ApproveIssueIntakeInput {
  factory Input$ApproveIssueIntakeInput({
    required String runID,
    required String taskID,
    required String jobID,
    required String projectID,
    required String source,
    required String issueReference,
    required String approvedBy,
  }) => Input$ApproveIssueIntakeInput._({
    r'runID': runID,
    r'taskID': taskID,
    r'jobID': jobID,
    r'projectID': projectID,
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
    final l$projectID = data['projectID'];
    result$data['projectID'] = (l$projectID as String);
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

  String get projectID => (_$data['projectID'] as String);

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
    final l$projectID = projectID;
    result$data['projectID'] = l$projectID;
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
    final l$projectID = projectID;
    final lOther$projectID = other.projectID;
    if (l$projectID != lOther$projectID) {
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
    final l$projectID = projectID;
    final l$source = source;
    final l$issueReference = issueReference;
    final l$approvedBy = approvedBy;
    return Object.hashAll([
      l$runID,
      l$taskID,
      l$jobID,
      l$projectID,
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
    String? projectID,
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
    Object? projectID = _undefined,
    Object? source = _undefined,
    Object? issueReference = _undefined,
    Object? approvedBy = _undefined,
  }) => _then(
    Input$ApproveIssueIntakeInput._({
      ..._instance._$data,
      if (runID != _undefined && runID != null) 'runID': (runID as String),
      if (taskID != _undefined && taskID != null) 'taskID': (taskID as String),
      if (jobID != _undefined && jobID != null) 'jobID': (jobID as String),
      if (projectID != _undefined && projectID != null)
        'projectID': (projectID as String),
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
    String? projectID,
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

class Input$ProjectRepositoryInput {
  factory Input$ProjectRepositoryInput({
    required String repositoryID,
    required String scmID,
    required String repositoryURL,
    required bool isPrimary,
  }) => Input$ProjectRepositoryInput._({
    r'repositoryID': repositoryID,
    r'scmID': scmID,
    r'repositoryURL': repositoryURL,
    r'isPrimary': isPrimary,
  });

  Input$ProjectRepositoryInput._(this._$data);

  factory Input$ProjectRepositoryInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$repositoryID = data['repositoryID'];
    result$data['repositoryID'] = (l$repositoryID as String);
    final l$scmID = data['scmID'];
    result$data['scmID'] = (l$scmID as String);
    final l$repositoryURL = data['repositoryURL'];
    result$data['repositoryURL'] = (l$repositoryURL as String);
    final l$isPrimary = data['isPrimary'];
    result$data['isPrimary'] = (l$isPrimary as bool);
    return Input$ProjectRepositoryInput._(result$data);
  }

  Map<String, dynamic> _$data;

  String get repositoryID => (_$data['repositoryID'] as String);

  String get scmID => (_$data['scmID'] as String);

  String get repositoryURL => (_$data['repositoryURL'] as String);

  bool get isPrimary => (_$data['isPrimary'] as bool);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$repositoryID = repositoryID;
    result$data['repositoryID'] = l$repositoryID;
    final l$scmID = scmID;
    result$data['scmID'] = l$scmID;
    final l$repositoryURL = repositoryURL;
    result$data['repositoryURL'] = l$repositoryURL;
    final l$isPrimary = isPrimary;
    result$data['isPrimary'] = l$isPrimary;
    return result$data;
  }

  CopyWith$Input$ProjectRepositoryInput<Input$ProjectRepositoryInput>
  get copyWith => CopyWith$Input$ProjectRepositoryInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$ProjectRepositoryInput ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$repositoryID = repositoryID;
    final lOther$repositoryID = other.repositoryID;
    if (l$repositoryID != lOther$repositoryID) {
      return false;
    }
    final l$scmID = scmID;
    final lOther$scmID = other.scmID;
    if (l$scmID != lOther$scmID) {
      return false;
    }
    final l$repositoryURL = repositoryURL;
    final lOther$repositoryURL = other.repositoryURL;
    if (l$repositoryURL != lOther$repositoryURL) {
      return false;
    }
    final l$isPrimary = isPrimary;
    final lOther$isPrimary = other.isPrimary;
    if (l$isPrimary != lOther$isPrimary) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$repositoryID = repositoryID;
    final l$scmID = scmID;
    final l$repositoryURL = repositoryURL;
    final l$isPrimary = isPrimary;
    return Object.hashAll([
      l$repositoryID,
      l$scmID,
      l$repositoryURL,
      l$isPrimary,
    ]);
  }
}

abstract class CopyWith$Input$ProjectRepositoryInput<TRes> {
  factory CopyWith$Input$ProjectRepositoryInput(
    Input$ProjectRepositoryInput instance,
    TRes Function(Input$ProjectRepositoryInput) then,
  ) = _CopyWithImpl$Input$ProjectRepositoryInput;

  factory CopyWith$Input$ProjectRepositoryInput.stub(TRes res) =
      _CopyWithStubImpl$Input$ProjectRepositoryInput;

  TRes call({
    String? repositoryID,
    String? scmID,
    String? repositoryURL,
    bool? isPrimary,
  });
}

class _CopyWithImpl$Input$ProjectRepositoryInput<TRes>
    implements CopyWith$Input$ProjectRepositoryInput<TRes> {
  _CopyWithImpl$Input$ProjectRepositoryInput(this._instance, this._then);

  final Input$ProjectRepositoryInput _instance;

  final TRes Function(Input$ProjectRepositoryInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? repositoryID = _undefined,
    Object? scmID = _undefined,
    Object? repositoryURL = _undefined,
    Object? isPrimary = _undefined,
  }) => _then(
    Input$ProjectRepositoryInput._({
      ..._instance._$data,
      if (repositoryID != _undefined && repositoryID != null)
        'repositoryID': (repositoryID as String),
      if (scmID != _undefined && scmID != null) 'scmID': (scmID as String),
      if (repositoryURL != _undefined && repositoryURL != null)
        'repositoryURL': (repositoryURL as String),
      if (isPrimary != _undefined && isPrimary != null)
        'isPrimary': (isPrimary as bool),
    }),
  );
}

class _CopyWithStubImpl$Input$ProjectRepositoryInput<TRes>
    implements CopyWith$Input$ProjectRepositoryInput<TRes> {
  _CopyWithStubImpl$Input$ProjectRepositoryInput(this._res);

  TRes _res;

  call({
    String? repositoryID,
    String? scmID,
    String? repositoryURL,
    bool? isPrimary,
  }) => _res;
}

class Input$ProjectSCMInput {
  factory Input$ProjectSCMInput({
    required String scmID,
    required Enum$SCMProvider scmProvider,
    required String scmToken,
  }) => Input$ProjectSCMInput._({
    r'scmID': scmID,
    r'scmProvider': scmProvider,
    r'scmToken': scmToken,
  });

  Input$ProjectSCMInput._(this._$data);

  factory Input$ProjectSCMInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$scmID = data['scmID'];
    result$data['scmID'] = (l$scmID as String);
    final l$scmProvider = data['scmProvider'];
    result$data['scmProvider'] = fromJson$Enum$SCMProvider(
      (l$scmProvider as String),
    );
    final l$scmToken = data['scmToken'];
    result$data['scmToken'] = (l$scmToken as String);
    return Input$ProjectSCMInput._(result$data);
  }

  Map<String, dynamic> _$data;

  String get scmID => (_$data['scmID'] as String);

  Enum$SCMProvider get scmProvider =>
      (_$data['scmProvider'] as Enum$SCMProvider);

  String get scmToken => (_$data['scmToken'] as String);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$scmID = scmID;
    result$data['scmID'] = l$scmID;
    final l$scmProvider = scmProvider;
    result$data['scmProvider'] = toJson$Enum$SCMProvider(l$scmProvider);
    final l$scmToken = scmToken;
    result$data['scmToken'] = l$scmToken;
    return result$data;
  }

  CopyWith$Input$ProjectSCMInput<Input$ProjectSCMInput> get copyWith =>
      CopyWith$Input$ProjectSCMInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$ProjectSCMInput || runtimeType != other.runtimeType) {
      return false;
    }
    final l$scmID = scmID;
    final lOther$scmID = other.scmID;
    if (l$scmID != lOther$scmID) {
      return false;
    }
    final l$scmProvider = scmProvider;
    final lOther$scmProvider = other.scmProvider;
    if (l$scmProvider != lOther$scmProvider) {
      return false;
    }
    final l$scmToken = scmToken;
    final lOther$scmToken = other.scmToken;
    if (l$scmToken != lOther$scmToken) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$scmID = scmID;
    final l$scmProvider = scmProvider;
    final l$scmToken = scmToken;
    return Object.hashAll([l$scmID, l$scmProvider, l$scmToken]);
  }
}

abstract class CopyWith$Input$ProjectSCMInput<TRes> {
  factory CopyWith$Input$ProjectSCMInput(
    Input$ProjectSCMInput instance,
    TRes Function(Input$ProjectSCMInput) then,
  ) = _CopyWithImpl$Input$ProjectSCMInput;

  factory CopyWith$Input$ProjectSCMInput.stub(TRes res) =
      _CopyWithStubImpl$Input$ProjectSCMInput;

  TRes call({String? scmID, Enum$SCMProvider? scmProvider, String? scmToken});
}

class _CopyWithImpl$Input$ProjectSCMInput<TRes>
    implements CopyWith$Input$ProjectSCMInput<TRes> {
  _CopyWithImpl$Input$ProjectSCMInput(this._instance, this._then);

  final Input$ProjectSCMInput _instance;

  final TRes Function(Input$ProjectSCMInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? scmID = _undefined,
    Object? scmProvider = _undefined,
    Object? scmToken = _undefined,
  }) => _then(
    Input$ProjectSCMInput._({
      ..._instance._$data,
      if (scmID != _undefined && scmID != null) 'scmID': (scmID as String),
      if (scmProvider != _undefined && scmProvider != null)
        'scmProvider': (scmProvider as Enum$SCMProvider),
      if (scmToken != _undefined && scmToken != null)
        'scmToken': (scmToken as String),
    }),
  );
}

class _CopyWithStubImpl$Input$ProjectSCMInput<TRes>
    implements CopyWith$Input$ProjectSCMInput<TRes> {
  _CopyWithStubImpl$Input$ProjectSCMInput(this._res);

  TRes _res;

  call({String? scmID, Enum$SCMProvider? scmProvider, String? scmToken}) =>
      _res;
}

class Input$ProjectBoardInput {
  factory Input$ProjectBoardInput({
    required Enum$TrackerSourceKind trackerProvider,
    String? taskboardName,
    required bool appliesToAllRepositories,
    List<String>? repositoryIDs,
  }) => Input$ProjectBoardInput._({
    r'trackerProvider': trackerProvider,
    if (taskboardName != null) r'taskboardName': taskboardName,
    r'appliesToAllRepositories': appliesToAllRepositories,
    if (repositoryIDs != null) r'repositoryIDs': repositoryIDs,
  });

  Input$ProjectBoardInput._(this._$data);

  factory Input$ProjectBoardInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$trackerProvider = data['trackerProvider'];
    result$data['trackerProvider'] = fromJson$Enum$TrackerSourceKind(
      (l$trackerProvider as String),
    );
    if (data.containsKey('taskboardName')) {
      final l$taskboardName = data['taskboardName'];
      result$data['taskboardName'] = (l$taskboardName as String?);
    }
    final l$appliesToAllRepositories = data['appliesToAllRepositories'];
    result$data['appliesToAllRepositories'] =
        (l$appliesToAllRepositories as bool);
    if (data.containsKey('repositoryIDs')) {
      final l$repositoryIDs = data['repositoryIDs'];
      result$data['repositoryIDs'] = (l$repositoryIDs as List<dynamic>?)
          ?.map((e) => (e as String))
          .toList();
    }
    return Input$ProjectBoardInput._(result$data);
  }

  Map<String, dynamic> _$data;

  Enum$TrackerSourceKind get trackerProvider =>
      (_$data['trackerProvider'] as Enum$TrackerSourceKind);

  String? get taskboardName => (_$data['taskboardName'] as String?);

  bool get appliesToAllRepositories =>
      (_$data['appliesToAllRepositories'] as bool);

  List<String>? get repositoryIDs => (_$data['repositoryIDs'] as List<String>?);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$trackerProvider = trackerProvider;
    result$data['trackerProvider'] = toJson$Enum$TrackerSourceKind(
      l$trackerProvider,
    );
    if (_$data.containsKey('taskboardName')) {
      final l$taskboardName = taskboardName;
      result$data['taskboardName'] = l$taskboardName;
    }
    final l$appliesToAllRepositories = appliesToAllRepositories;
    result$data['appliesToAllRepositories'] = l$appliesToAllRepositories;
    if (_$data.containsKey('repositoryIDs')) {
      final l$repositoryIDs = repositoryIDs;
      result$data['repositoryIDs'] = l$repositoryIDs?.map((e) => e).toList();
    }
    return result$data;
  }

  CopyWith$Input$ProjectBoardInput<Input$ProjectBoardInput> get copyWith =>
      CopyWith$Input$ProjectBoardInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$ProjectBoardInput || runtimeType != other.runtimeType) {
      return false;
    }
    final l$trackerProvider = trackerProvider;
    final lOther$trackerProvider = other.trackerProvider;
    if (l$trackerProvider != lOther$trackerProvider) {
      return false;
    }
    final l$taskboardName = taskboardName;
    final lOther$taskboardName = other.taskboardName;
    if (_$data.containsKey('taskboardName') !=
        other._$data.containsKey('taskboardName')) {
      return false;
    }
    if (l$taskboardName != lOther$taskboardName) {
      return false;
    }
    final l$appliesToAllRepositories = appliesToAllRepositories;
    final lOther$appliesToAllRepositories = other.appliesToAllRepositories;
    if (l$appliesToAllRepositories != lOther$appliesToAllRepositories) {
      return false;
    }
    final l$repositoryIDs = repositoryIDs;
    final lOther$repositoryIDs = other.repositoryIDs;
    if (_$data.containsKey('repositoryIDs') !=
        other._$data.containsKey('repositoryIDs')) {
      return false;
    }
    if (l$repositoryIDs != null && lOther$repositoryIDs != null) {
      if (l$repositoryIDs.length != lOther$repositoryIDs.length) {
        return false;
      }
      for (int i = 0; i < l$repositoryIDs.length; i++) {
        final l$repositoryIDs$entry = l$repositoryIDs[i];
        final lOther$repositoryIDs$entry = lOther$repositoryIDs[i];
        if (l$repositoryIDs$entry != lOther$repositoryIDs$entry) {
          return false;
        }
      }
    } else if (l$repositoryIDs != lOther$repositoryIDs) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$trackerProvider = trackerProvider;
    final l$taskboardName = taskboardName;
    final l$appliesToAllRepositories = appliesToAllRepositories;
    final l$repositoryIDs = repositoryIDs;
    return Object.hashAll([
      l$trackerProvider,
      _$data.containsKey('taskboardName') ? l$taskboardName : const {},
      l$appliesToAllRepositories,
      _$data.containsKey('repositoryIDs')
          ? l$repositoryIDs == null
                ? null
                : Object.hashAll(l$repositoryIDs.map((v) => v))
          : const {},
    ]);
  }
}

abstract class CopyWith$Input$ProjectBoardInput<TRes> {
  factory CopyWith$Input$ProjectBoardInput(
    Input$ProjectBoardInput instance,
    TRes Function(Input$ProjectBoardInput) then,
  ) = _CopyWithImpl$Input$ProjectBoardInput;

  factory CopyWith$Input$ProjectBoardInput.stub(TRes res) =
      _CopyWithStubImpl$Input$ProjectBoardInput;

  TRes call({
    Enum$TrackerSourceKind? trackerProvider,
    String? taskboardName,
    bool? appliesToAllRepositories,
    List<String>? repositoryIDs,
  });
}

class _CopyWithImpl$Input$ProjectBoardInput<TRes>
    implements CopyWith$Input$ProjectBoardInput<TRes> {
  _CopyWithImpl$Input$ProjectBoardInput(this._instance, this._then);

  final Input$ProjectBoardInput _instance;

  final TRes Function(Input$ProjectBoardInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? trackerProvider = _undefined,
    Object? taskboardName = _undefined,
    Object? appliesToAllRepositories = _undefined,
    Object? repositoryIDs = _undefined,
  }) => _then(
    Input$ProjectBoardInput._({
      ..._instance._$data,
      if (trackerProvider != _undefined && trackerProvider != null)
        'trackerProvider': (trackerProvider as Enum$TrackerSourceKind),
      if (taskboardName != _undefined)
        'taskboardName': (taskboardName as String?),
      if (appliesToAllRepositories != _undefined &&
          appliesToAllRepositories != null)
        'appliesToAllRepositories': (appliesToAllRepositories as bool),
      if (repositoryIDs != _undefined)
        'repositoryIDs': (repositoryIDs as List<String>?),
    }),
  );
}

class _CopyWithStubImpl$Input$ProjectBoardInput<TRes>
    implements CopyWith$Input$ProjectBoardInput<TRes> {
  _CopyWithStubImpl$Input$ProjectBoardInput(this._res);

  TRes _res;

  call({
    Enum$TrackerSourceKind? trackerProvider,
    String? taskboardName,
    bool? appliesToAllRepositories,
    List<String>? repositoryIDs,
  }) => _res;
}

class Input$UpsertProjectSetupInput {
  factory Input$UpsertProjectSetupInput({
    required String projectID,
    required String projectName,
    required List<Input$ProjectSCMInput> scms,
    required List<Input$ProjectRepositoryInput> repositories,
    required List<Input$ProjectBoardInput> boards,
  }) => Input$UpsertProjectSetupInput._({
    r'projectID': projectID,
    r'projectName': projectName,
    r'scms': scms,
    r'repositories': repositories,
    r'boards': boards,
  });

  Input$UpsertProjectSetupInput._(this._$data);

  factory Input$UpsertProjectSetupInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$projectID = data['projectID'];
    result$data['projectID'] = (l$projectID as String);
    final l$projectName = data['projectName'];
    result$data['projectName'] = (l$projectName as String);
    final l$scms = data['scms'];
    result$data['scms'] = (l$scms as List<dynamic>)
        .map((e) => Input$ProjectSCMInput.fromJson((e as Map<String, dynamic>)))
        .toList();
    final l$repositories = data['repositories'];
    result$data['repositories'] = (l$repositories as List<dynamic>)
        .map(
          (e) => Input$ProjectRepositoryInput.fromJson(
            (e as Map<String, dynamic>),
          ),
        )
        .toList();
    final l$boards = data['boards'];
    result$data['boards'] = (l$boards as List<dynamic>)
        .map(
          (e) => Input$ProjectBoardInput.fromJson((e as Map<String, dynamic>)),
        )
        .toList();
    return Input$UpsertProjectSetupInput._(result$data);
  }

  Map<String, dynamic> _$data;

  String get projectID => (_$data['projectID'] as String);

  String get projectName => (_$data['projectName'] as String);

  List<Input$ProjectSCMInput> get scms =>
      (_$data['scms'] as List<Input$ProjectSCMInput>);

  List<Input$ProjectRepositoryInput> get repositories =>
      (_$data['repositories'] as List<Input$ProjectRepositoryInput>);

  List<Input$ProjectBoardInput> get boards =>
      (_$data['boards'] as List<Input$ProjectBoardInput>);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$projectID = projectID;
    result$data['projectID'] = l$projectID;
    final l$projectName = projectName;
    result$data['projectName'] = l$projectName;
    final l$scms = scms;
    result$data['scms'] = l$scms.map((e) => e.toJson()).toList();
    final l$repositories = repositories;
    result$data['repositories'] = l$repositories
        .map((e) => e.toJson())
        .toList();
    final l$boards = boards;
    result$data['boards'] = l$boards.map((e) => e.toJson()).toList();
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
    final l$scms = scms;
    final lOther$scms = other.scms;
    if (l$scms.length != lOther$scms.length) {
      return false;
    }
    for (int i = 0; i < l$scms.length; i++) {
      final l$scms$entry = l$scms[i];
      final lOther$scms$entry = lOther$scms[i];
      if (l$scms$entry != lOther$scms$entry) {
        return false;
      }
    }
    final l$repositories = repositories;
    final lOther$repositories = other.repositories;
    if (l$repositories.length != lOther$repositories.length) {
      return false;
    }
    for (int i = 0; i < l$repositories.length; i++) {
      final l$repositories$entry = l$repositories[i];
      final lOther$repositories$entry = lOther$repositories[i];
      if (l$repositories$entry != lOther$repositories$entry) {
        return false;
      }
    }
    final l$boards = boards;
    final lOther$boards = other.boards;
    if (l$boards.length != lOther$boards.length) {
      return false;
    }
    for (int i = 0; i < l$boards.length; i++) {
      final l$boards$entry = l$boards[i];
      final lOther$boards$entry = lOther$boards[i];
      if (l$boards$entry != lOther$boards$entry) {
        return false;
      }
    }
    return true;
  }

  @override
  int get hashCode {
    final l$projectID = projectID;
    final l$projectName = projectName;
    final l$scms = scms;
    final l$repositories = repositories;
    final l$boards = boards;
    return Object.hashAll([
      l$projectID,
      l$projectName,
      Object.hashAll(l$scms.map((v) => v)),
      Object.hashAll(l$repositories.map((v) => v)),
      Object.hashAll(l$boards.map((v) => v)),
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
    List<Input$ProjectSCMInput>? scms,
    List<Input$ProjectRepositoryInput>? repositories,
    List<Input$ProjectBoardInput>? boards,
  });
  TRes scms(
    Iterable<Input$ProjectSCMInput> Function(
      Iterable<CopyWith$Input$ProjectSCMInput<Input$ProjectSCMInput>>,
    )
    _fn,
  );
  TRes repositories(
    Iterable<Input$ProjectRepositoryInput> Function(
      Iterable<
        CopyWith$Input$ProjectRepositoryInput<Input$ProjectRepositoryInput>
      >,
    )
    _fn,
  );
  TRes boards(
    Iterable<Input$ProjectBoardInput> Function(
      Iterable<CopyWith$Input$ProjectBoardInput<Input$ProjectBoardInput>>,
    )
    _fn,
  );
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
    Object? scms = _undefined,
    Object? repositories = _undefined,
    Object? boards = _undefined,
  }) => _then(
    Input$UpsertProjectSetupInput._({
      ..._instance._$data,
      if (projectID != _undefined && projectID != null)
        'projectID': (projectID as String),
      if (projectName != _undefined && projectName != null)
        'projectName': (projectName as String),
      if (scms != _undefined && scms != null)
        'scms': (scms as List<Input$ProjectSCMInput>),
      if (repositories != _undefined && repositories != null)
        'repositories': (repositories as List<Input$ProjectRepositoryInput>),
      if (boards != _undefined && boards != null)
        'boards': (boards as List<Input$ProjectBoardInput>),
    }),
  );

  TRes scms(
    Iterable<Input$ProjectSCMInput> Function(
      Iterable<CopyWith$Input$ProjectSCMInput<Input$ProjectSCMInput>>,
    )
    _fn,
  ) => call(
    scms: _fn(
      _instance.scms.map((e) => CopyWith$Input$ProjectSCMInput(e, (i) => i)),
    ).toList(),
  );

  TRes repositories(
    Iterable<Input$ProjectRepositoryInput> Function(
      Iterable<
        CopyWith$Input$ProjectRepositoryInput<Input$ProjectRepositoryInput>
      >,
    )
    _fn,
  ) => call(
    repositories: _fn(
      _instance.repositories.map(
        (e) => CopyWith$Input$ProjectRepositoryInput(e, (i) => i),
      ),
    ).toList(),
  );

  TRes boards(
    Iterable<Input$ProjectBoardInput> Function(
      Iterable<CopyWith$Input$ProjectBoardInput<Input$ProjectBoardInput>>,
    )
    _fn,
  ) => call(
    boards: _fn(
      _instance.boards.map(
        (e) => CopyWith$Input$ProjectBoardInput(e, (i) => i),
      ),
    ).toList(),
  );
}

class _CopyWithStubImpl$Input$UpsertProjectSetupInput<TRes>
    implements CopyWith$Input$UpsertProjectSetupInput<TRes> {
  _CopyWithStubImpl$Input$UpsertProjectSetupInput(this._res);

  TRes _res;

  call({
    String? projectID,
    String? projectName,
    List<Input$ProjectSCMInput>? scms,
    List<Input$ProjectRepositoryInput>? repositories,
    List<Input$ProjectBoardInput>? boards,
  }) => _res;

  scms(_fn) => _res;

  repositories(_fn) => _res;

  boards(_fn) => _res;
}

class Input$UpdateWorkerSettingsInput {
  factory Input$UpdateWorkerSettingsInput({
    required int heartbeatIntervalSeconds,
    required int responseDeadlineSeconds,
    required int staleAfterSeconds,
    required int drainTimeoutSeconds,
    required int terminateTimeoutSeconds,
    required int rogueThreshold,
  }) => Input$UpdateWorkerSettingsInput._({
    r'heartbeatIntervalSeconds': heartbeatIntervalSeconds,
    r'responseDeadlineSeconds': responseDeadlineSeconds,
    r'staleAfterSeconds': staleAfterSeconds,
    r'drainTimeoutSeconds': drainTimeoutSeconds,
    r'terminateTimeoutSeconds': terminateTimeoutSeconds,
    r'rogueThreshold': rogueThreshold,
  });

  Input$UpdateWorkerSettingsInput._(this._$data);

  factory Input$UpdateWorkerSettingsInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$heartbeatIntervalSeconds = data['heartbeatIntervalSeconds'];
    result$data['heartbeatIntervalSeconds'] =
        (l$heartbeatIntervalSeconds as int);
    final l$responseDeadlineSeconds = data['responseDeadlineSeconds'];
    result$data['responseDeadlineSeconds'] = (l$responseDeadlineSeconds as int);
    final l$staleAfterSeconds = data['staleAfterSeconds'];
    result$data['staleAfterSeconds'] = (l$staleAfterSeconds as int);
    final l$drainTimeoutSeconds = data['drainTimeoutSeconds'];
    result$data['drainTimeoutSeconds'] = (l$drainTimeoutSeconds as int);
    final l$terminateTimeoutSeconds = data['terminateTimeoutSeconds'];
    result$data['terminateTimeoutSeconds'] = (l$terminateTimeoutSeconds as int);
    final l$rogueThreshold = data['rogueThreshold'];
    result$data['rogueThreshold'] = (l$rogueThreshold as int);
    return Input$UpdateWorkerSettingsInput._(result$data);
  }

  Map<String, dynamic> _$data;

  int get heartbeatIntervalSeconds =>
      (_$data['heartbeatIntervalSeconds'] as int);

  int get responseDeadlineSeconds => (_$data['responseDeadlineSeconds'] as int);

  int get staleAfterSeconds => (_$data['staleAfterSeconds'] as int);

  int get drainTimeoutSeconds => (_$data['drainTimeoutSeconds'] as int);

  int get terminateTimeoutSeconds => (_$data['terminateTimeoutSeconds'] as int);

  int get rogueThreshold => (_$data['rogueThreshold'] as int);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$heartbeatIntervalSeconds = heartbeatIntervalSeconds;
    result$data['heartbeatIntervalSeconds'] = l$heartbeatIntervalSeconds;
    final l$responseDeadlineSeconds = responseDeadlineSeconds;
    result$data['responseDeadlineSeconds'] = l$responseDeadlineSeconds;
    final l$staleAfterSeconds = staleAfterSeconds;
    result$data['staleAfterSeconds'] = l$staleAfterSeconds;
    final l$drainTimeoutSeconds = drainTimeoutSeconds;
    result$data['drainTimeoutSeconds'] = l$drainTimeoutSeconds;
    final l$terminateTimeoutSeconds = terminateTimeoutSeconds;
    result$data['terminateTimeoutSeconds'] = l$terminateTimeoutSeconds;
    final l$rogueThreshold = rogueThreshold;
    result$data['rogueThreshold'] = l$rogueThreshold;
    return result$data;
  }

  CopyWith$Input$UpdateWorkerSettingsInput<Input$UpdateWorkerSettingsInput>
  get copyWith => CopyWith$Input$UpdateWorkerSettingsInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$UpdateWorkerSettingsInput ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$heartbeatIntervalSeconds = heartbeatIntervalSeconds;
    final lOther$heartbeatIntervalSeconds = other.heartbeatIntervalSeconds;
    if (l$heartbeatIntervalSeconds != lOther$heartbeatIntervalSeconds) {
      return false;
    }
    final l$responseDeadlineSeconds = responseDeadlineSeconds;
    final lOther$responseDeadlineSeconds = other.responseDeadlineSeconds;
    if (l$responseDeadlineSeconds != lOther$responseDeadlineSeconds) {
      return false;
    }
    final l$staleAfterSeconds = staleAfterSeconds;
    final lOther$staleAfterSeconds = other.staleAfterSeconds;
    if (l$staleAfterSeconds != lOther$staleAfterSeconds) {
      return false;
    }
    final l$drainTimeoutSeconds = drainTimeoutSeconds;
    final lOther$drainTimeoutSeconds = other.drainTimeoutSeconds;
    if (l$drainTimeoutSeconds != lOther$drainTimeoutSeconds) {
      return false;
    }
    final l$terminateTimeoutSeconds = terminateTimeoutSeconds;
    final lOther$terminateTimeoutSeconds = other.terminateTimeoutSeconds;
    if (l$terminateTimeoutSeconds != lOther$terminateTimeoutSeconds) {
      return false;
    }
    final l$rogueThreshold = rogueThreshold;
    final lOther$rogueThreshold = other.rogueThreshold;
    if (l$rogueThreshold != lOther$rogueThreshold) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$heartbeatIntervalSeconds = heartbeatIntervalSeconds;
    final l$responseDeadlineSeconds = responseDeadlineSeconds;
    final l$staleAfterSeconds = staleAfterSeconds;
    final l$drainTimeoutSeconds = drainTimeoutSeconds;
    final l$terminateTimeoutSeconds = terminateTimeoutSeconds;
    final l$rogueThreshold = rogueThreshold;
    return Object.hashAll([
      l$heartbeatIntervalSeconds,
      l$responseDeadlineSeconds,
      l$staleAfterSeconds,
      l$drainTimeoutSeconds,
      l$terminateTimeoutSeconds,
      l$rogueThreshold,
    ]);
  }
}

abstract class CopyWith$Input$UpdateWorkerSettingsInput<TRes> {
  factory CopyWith$Input$UpdateWorkerSettingsInput(
    Input$UpdateWorkerSettingsInput instance,
    TRes Function(Input$UpdateWorkerSettingsInput) then,
  ) = _CopyWithImpl$Input$UpdateWorkerSettingsInput;

  factory CopyWith$Input$UpdateWorkerSettingsInput.stub(TRes res) =
      _CopyWithStubImpl$Input$UpdateWorkerSettingsInput;

  TRes call({
    int? heartbeatIntervalSeconds,
    int? responseDeadlineSeconds,
    int? staleAfterSeconds,
    int? drainTimeoutSeconds,
    int? terminateTimeoutSeconds,
    int? rogueThreshold,
  });
}

class _CopyWithImpl$Input$UpdateWorkerSettingsInput<TRes>
    implements CopyWith$Input$UpdateWorkerSettingsInput<TRes> {
  _CopyWithImpl$Input$UpdateWorkerSettingsInput(this._instance, this._then);

  final Input$UpdateWorkerSettingsInput _instance;

  final TRes Function(Input$UpdateWorkerSettingsInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? heartbeatIntervalSeconds = _undefined,
    Object? responseDeadlineSeconds = _undefined,
    Object? staleAfterSeconds = _undefined,
    Object? drainTimeoutSeconds = _undefined,
    Object? terminateTimeoutSeconds = _undefined,
    Object? rogueThreshold = _undefined,
  }) => _then(
    Input$UpdateWorkerSettingsInput._({
      ..._instance._$data,
      if (heartbeatIntervalSeconds != _undefined &&
          heartbeatIntervalSeconds != null)
        'heartbeatIntervalSeconds': (heartbeatIntervalSeconds as int),
      if (responseDeadlineSeconds != _undefined &&
          responseDeadlineSeconds != null)
        'responseDeadlineSeconds': (responseDeadlineSeconds as int),
      if (staleAfterSeconds != _undefined && staleAfterSeconds != null)
        'staleAfterSeconds': (staleAfterSeconds as int),
      if (drainTimeoutSeconds != _undefined && drainTimeoutSeconds != null)
        'drainTimeoutSeconds': (drainTimeoutSeconds as int),
      if (terminateTimeoutSeconds != _undefined &&
          terminateTimeoutSeconds != null)
        'terminateTimeoutSeconds': (terminateTimeoutSeconds as int),
      if (rogueThreshold != _undefined && rogueThreshold != null)
        'rogueThreshold': (rogueThreshold as int),
    }),
  );
}

class _CopyWithStubImpl$Input$UpdateWorkerSettingsInput<TRes>
    implements CopyWith$Input$UpdateWorkerSettingsInput<TRes> {
  _CopyWithStubImpl$Input$UpdateWorkerSettingsInput(this._res);

  TRes _res;

  call({
    int? heartbeatIntervalSeconds,
    int? responseDeadlineSeconds,
    int? staleAfterSeconds,
    int? drainTimeoutSeconds,
    int? terminateTimeoutSeconds,
    int? rogueThreshold,
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
  INTERNAL,
  GITHUB_ISSUES,
  $unknown;

  factory Enum$TrackerSourceKind.fromJson(String value) =>
      fromJson$Enum$TrackerSourceKind(value);

  String toJson() => toJson$Enum$TrackerSourceKind(this);
}

String toJson$Enum$TrackerSourceKind(Enum$TrackerSourceKind e) {
  switch (e) {
    case Enum$TrackerSourceKind.INTERNAL:
      return r'INTERNAL';
    case Enum$TrackerSourceKind.GITHUB_ISSUES:
      return r'GITHUB_ISSUES';
    case Enum$TrackerSourceKind.$unknown:
      return r'$unknown';
  }
}

Enum$TrackerSourceKind fromJson$Enum$TrackerSourceKind(String value) {
  switch (value) {
    case r'INTERNAL':
      return Enum$TrackerSourceKind.INTERNAL;
    case r'GITHUB_ISSUES':
      return Enum$TrackerSourceKind.GITHUB_ISSUES;
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
