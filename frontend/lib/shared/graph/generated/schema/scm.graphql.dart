class Input$EnqueueSCMWorkflowInput {
  factory Input$EnqueueSCMWorkflowInput({
    required Enum$SCMOperation operation,
    required Enum$SCMProvider provider,
    required String owner,
    required String repository,
    required String runID,
    required String taskID,
    required String jobID,
    required String idempotencyKey,
    String? worktreePath,
    String? baseBranch,
    String? targetBranch,
    int? pullRequestNumber,
    Enum$SCMMergeMethod? mergeMethod,
    String? pullRequestTitle,
    String? pullRequestBody,
    Enum$SCMReviewDecision? reviewDecision,
    String? reviewBody,
  }) => Input$EnqueueSCMWorkflowInput._({
    r'operation': operation,
    r'provider': provider,
    r'owner': owner,
    r'repository': repository,
    r'runID': runID,
    r'taskID': taskID,
    r'jobID': jobID,
    r'idempotencyKey': idempotencyKey,
    if (worktreePath != null) r'worktreePath': worktreePath,
    if (baseBranch != null) r'baseBranch': baseBranch,
    if (targetBranch != null) r'targetBranch': targetBranch,
    if (pullRequestNumber != null) r'pullRequestNumber': pullRequestNumber,
    if (mergeMethod != null) r'mergeMethod': mergeMethod,
    if (pullRequestTitle != null) r'pullRequestTitle': pullRequestTitle,
    if (pullRequestBody != null) r'pullRequestBody': pullRequestBody,
    if (reviewDecision != null) r'reviewDecision': reviewDecision,
    if (reviewBody != null) r'reviewBody': reviewBody,
  });

  Input$EnqueueSCMWorkflowInput._(this._$data);

  factory Input$EnqueueSCMWorkflowInput.fromJson(Map<String, dynamic> data) {
    final result$data = <String, dynamic>{};
    final l$operation = data['operation'];
    result$data['operation'] = fromJson$Enum$SCMOperation(
      (l$operation as String),
    );
    final l$provider = data['provider'];
    result$data['provider'] = fromJson$Enum$SCMProvider((l$provider as String));
    final l$owner = data['owner'];
    result$data['owner'] = (l$owner as String);
    final l$repository = data['repository'];
    result$data['repository'] = (l$repository as String);
    final l$runID = data['runID'];
    result$data['runID'] = (l$runID as String);
    final l$taskID = data['taskID'];
    result$data['taskID'] = (l$taskID as String);
    final l$jobID = data['jobID'];
    result$data['jobID'] = (l$jobID as String);
    final l$idempotencyKey = data['idempotencyKey'];
    result$data['idempotencyKey'] = (l$idempotencyKey as String);
    if (data.containsKey('worktreePath')) {
      final l$worktreePath = data['worktreePath'];
      result$data['worktreePath'] = (l$worktreePath as String?);
    }
    if (data.containsKey('baseBranch')) {
      final l$baseBranch = data['baseBranch'];
      result$data['baseBranch'] = (l$baseBranch as String?);
    }
    if (data.containsKey('targetBranch')) {
      final l$targetBranch = data['targetBranch'];
      result$data['targetBranch'] = (l$targetBranch as String?);
    }
    if (data.containsKey('pullRequestNumber')) {
      final l$pullRequestNumber = data['pullRequestNumber'];
      result$data['pullRequestNumber'] = (l$pullRequestNumber as int?);
    }
    if (data.containsKey('mergeMethod')) {
      final l$mergeMethod = data['mergeMethod'];
      result$data['mergeMethod'] = l$mergeMethod == null
          ? null
          : fromJson$Enum$SCMMergeMethod((l$mergeMethod as String));
    }
    if (data.containsKey('pullRequestTitle')) {
      final l$pullRequestTitle = data['pullRequestTitle'];
      result$data['pullRequestTitle'] = (l$pullRequestTitle as String?);
    }
    if (data.containsKey('pullRequestBody')) {
      final l$pullRequestBody = data['pullRequestBody'];
      result$data['pullRequestBody'] = (l$pullRequestBody as String?);
    }
    if (data.containsKey('reviewDecision')) {
      final l$reviewDecision = data['reviewDecision'];
      result$data['reviewDecision'] = l$reviewDecision == null
          ? null
          : fromJson$Enum$SCMReviewDecision((l$reviewDecision as String));
    }
    if (data.containsKey('reviewBody')) {
      final l$reviewBody = data['reviewBody'];
      result$data['reviewBody'] = (l$reviewBody as String?);
    }
    return Input$EnqueueSCMWorkflowInput._(result$data);
  }

  Map<String, dynamic> _$data;

  Enum$SCMOperation get operation => (_$data['operation'] as Enum$SCMOperation);

  Enum$SCMProvider get provider => (_$data['provider'] as Enum$SCMProvider);

  String get owner => (_$data['owner'] as String);

  String get repository => (_$data['repository'] as String);

  String get runID => (_$data['runID'] as String);

  String get taskID => (_$data['taskID'] as String);

  String get jobID => (_$data['jobID'] as String);

  String get idempotencyKey => (_$data['idempotencyKey'] as String);

  String? get worktreePath => (_$data['worktreePath'] as String?);

  String? get baseBranch => (_$data['baseBranch'] as String?);

  String? get targetBranch => (_$data['targetBranch'] as String?);

  int? get pullRequestNumber => (_$data['pullRequestNumber'] as int?);

  Enum$SCMMergeMethod? get mergeMethod =>
      (_$data['mergeMethod'] as Enum$SCMMergeMethod?);

  String? get pullRequestTitle => (_$data['pullRequestTitle'] as String?);

  String? get pullRequestBody => (_$data['pullRequestBody'] as String?);

  Enum$SCMReviewDecision? get reviewDecision =>
      (_$data['reviewDecision'] as Enum$SCMReviewDecision?);

  String? get reviewBody => (_$data['reviewBody'] as String?);

  Map<String, dynamic> toJson() {
    final result$data = <String, dynamic>{};
    final l$operation = operation;
    result$data['operation'] = toJson$Enum$SCMOperation(l$operation);
    final l$provider = provider;
    result$data['provider'] = toJson$Enum$SCMProvider(l$provider);
    final l$owner = owner;
    result$data['owner'] = l$owner;
    final l$repository = repository;
    result$data['repository'] = l$repository;
    final l$runID = runID;
    result$data['runID'] = l$runID;
    final l$taskID = taskID;
    result$data['taskID'] = l$taskID;
    final l$jobID = jobID;
    result$data['jobID'] = l$jobID;
    final l$idempotencyKey = idempotencyKey;
    result$data['idempotencyKey'] = l$idempotencyKey;
    if (_$data.containsKey('worktreePath')) {
      final l$worktreePath = worktreePath;
      result$data['worktreePath'] = l$worktreePath;
    }
    if (_$data.containsKey('baseBranch')) {
      final l$baseBranch = baseBranch;
      result$data['baseBranch'] = l$baseBranch;
    }
    if (_$data.containsKey('targetBranch')) {
      final l$targetBranch = targetBranch;
      result$data['targetBranch'] = l$targetBranch;
    }
    if (_$data.containsKey('pullRequestNumber')) {
      final l$pullRequestNumber = pullRequestNumber;
      result$data['pullRequestNumber'] = l$pullRequestNumber;
    }
    if (_$data.containsKey('mergeMethod')) {
      final l$mergeMethod = mergeMethod;
      result$data['mergeMethod'] = l$mergeMethod == null
          ? null
          : toJson$Enum$SCMMergeMethod(l$mergeMethod);
    }
    if (_$data.containsKey('pullRequestTitle')) {
      final l$pullRequestTitle = pullRequestTitle;
      result$data['pullRequestTitle'] = l$pullRequestTitle;
    }
    if (_$data.containsKey('pullRequestBody')) {
      final l$pullRequestBody = pullRequestBody;
      result$data['pullRequestBody'] = l$pullRequestBody;
    }
    if (_$data.containsKey('reviewDecision')) {
      final l$reviewDecision = reviewDecision;
      result$data['reviewDecision'] = l$reviewDecision == null
          ? null
          : toJson$Enum$SCMReviewDecision(l$reviewDecision);
    }
    if (_$data.containsKey('reviewBody')) {
      final l$reviewBody = reviewBody;
      result$data['reviewBody'] = l$reviewBody;
    }
    return result$data;
  }

  CopyWith$Input$EnqueueSCMWorkflowInput<Input$EnqueueSCMWorkflowInput>
  get copyWith => CopyWith$Input$EnqueueSCMWorkflowInput(this, (i) => i);

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) {
      return true;
    }
    if (other is! Input$EnqueueSCMWorkflowInput ||
        runtimeType != other.runtimeType) {
      return false;
    }
    final l$operation = operation;
    final lOther$operation = other.operation;
    if (l$operation != lOther$operation) {
      return false;
    }
    final l$provider = provider;
    final lOther$provider = other.provider;
    if (l$provider != lOther$provider) {
      return false;
    }
    final l$owner = owner;
    final lOther$owner = other.owner;
    if (l$owner != lOther$owner) {
      return false;
    }
    final l$repository = repository;
    final lOther$repository = other.repository;
    if (l$repository != lOther$repository) {
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
    final l$worktreePath = worktreePath;
    final lOther$worktreePath = other.worktreePath;
    if (_$data.containsKey('worktreePath') !=
        other._$data.containsKey('worktreePath')) {
      return false;
    }
    if (l$worktreePath != lOther$worktreePath) {
      return false;
    }
    final l$baseBranch = baseBranch;
    final lOther$baseBranch = other.baseBranch;
    if (_$data.containsKey('baseBranch') !=
        other._$data.containsKey('baseBranch')) {
      return false;
    }
    if (l$baseBranch != lOther$baseBranch) {
      return false;
    }
    final l$targetBranch = targetBranch;
    final lOther$targetBranch = other.targetBranch;
    if (_$data.containsKey('targetBranch') !=
        other._$data.containsKey('targetBranch')) {
      return false;
    }
    if (l$targetBranch != lOther$targetBranch) {
      return false;
    }
    final l$pullRequestNumber = pullRequestNumber;
    final lOther$pullRequestNumber = other.pullRequestNumber;
    if (_$data.containsKey('pullRequestNumber') !=
        other._$data.containsKey('pullRequestNumber')) {
      return false;
    }
    if (l$pullRequestNumber != lOther$pullRequestNumber) {
      return false;
    }
    final l$mergeMethod = mergeMethod;
    final lOther$mergeMethod = other.mergeMethod;
    if (_$data.containsKey('mergeMethod') !=
        other._$data.containsKey('mergeMethod')) {
      return false;
    }
    if (l$mergeMethod != lOther$mergeMethod) {
      return false;
    }
    final l$pullRequestTitle = pullRequestTitle;
    final lOther$pullRequestTitle = other.pullRequestTitle;
    if (_$data.containsKey('pullRequestTitle') !=
        other._$data.containsKey('pullRequestTitle')) {
      return false;
    }
    if (l$pullRequestTitle != lOther$pullRequestTitle) {
      return false;
    }
    final l$pullRequestBody = pullRequestBody;
    final lOther$pullRequestBody = other.pullRequestBody;
    if (_$data.containsKey('pullRequestBody') !=
        other._$data.containsKey('pullRequestBody')) {
      return false;
    }
    if (l$pullRequestBody != lOther$pullRequestBody) {
      return false;
    }
    final l$reviewDecision = reviewDecision;
    final lOther$reviewDecision = other.reviewDecision;
    if (_$data.containsKey('reviewDecision') !=
        other._$data.containsKey('reviewDecision')) {
      return false;
    }
    if (l$reviewDecision != lOther$reviewDecision) {
      return false;
    }
    final l$reviewBody = reviewBody;
    final lOther$reviewBody = other.reviewBody;
    if (_$data.containsKey('reviewBody') !=
        other._$data.containsKey('reviewBody')) {
      return false;
    }
    if (l$reviewBody != lOther$reviewBody) {
      return false;
    }
    return true;
  }

  @override
  int get hashCode {
    final l$operation = operation;
    final l$provider = provider;
    final l$owner = owner;
    final l$repository = repository;
    final l$runID = runID;
    final l$taskID = taskID;
    final l$jobID = jobID;
    final l$idempotencyKey = idempotencyKey;
    final l$worktreePath = worktreePath;
    final l$baseBranch = baseBranch;
    final l$targetBranch = targetBranch;
    final l$pullRequestNumber = pullRequestNumber;
    final l$mergeMethod = mergeMethod;
    final l$pullRequestTitle = pullRequestTitle;
    final l$pullRequestBody = pullRequestBody;
    final l$reviewDecision = reviewDecision;
    final l$reviewBody = reviewBody;
    return Object.hashAll([
      l$operation,
      l$provider,
      l$owner,
      l$repository,
      l$runID,
      l$taskID,
      l$jobID,
      l$idempotencyKey,
      _$data.containsKey('worktreePath') ? l$worktreePath : const {},
      _$data.containsKey('baseBranch') ? l$baseBranch : const {},
      _$data.containsKey('targetBranch') ? l$targetBranch : const {},
      _$data.containsKey('pullRequestNumber') ? l$pullRequestNumber : const {},
      _$data.containsKey('mergeMethod') ? l$mergeMethod : const {},
      _$data.containsKey('pullRequestTitle') ? l$pullRequestTitle : const {},
      _$data.containsKey('pullRequestBody') ? l$pullRequestBody : const {},
      _$data.containsKey('reviewDecision') ? l$reviewDecision : const {},
      _$data.containsKey('reviewBody') ? l$reviewBody : const {},
    ]);
  }
}

abstract class CopyWith$Input$EnqueueSCMWorkflowInput<TRes> {
  factory CopyWith$Input$EnqueueSCMWorkflowInput(
    Input$EnqueueSCMWorkflowInput instance,
    TRes Function(Input$EnqueueSCMWorkflowInput) then,
  ) = _CopyWithImpl$Input$EnqueueSCMWorkflowInput;

  factory CopyWith$Input$EnqueueSCMWorkflowInput.stub(TRes res) =
      _CopyWithStubImpl$Input$EnqueueSCMWorkflowInput;

  TRes call({
    Enum$SCMOperation? operation,
    Enum$SCMProvider? provider,
    String? owner,
    String? repository,
    String? runID,
    String? taskID,
    String? jobID,
    String? idempotencyKey,
    String? worktreePath,
    String? baseBranch,
    String? targetBranch,
    int? pullRequestNumber,
    Enum$SCMMergeMethod? mergeMethod,
    String? pullRequestTitle,
    String? pullRequestBody,
    Enum$SCMReviewDecision? reviewDecision,
    String? reviewBody,
  });
}

class _CopyWithImpl$Input$EnqueueSCMWorkflowInput<TRes>
    implements CopyWith$Input$EnqueueSCMWorkflowInput<TRes> {
  _CopyWithImpl$Input$EnqueueSCMWorkflowInput(this._instance, this._then);

  final Input$EnqueueSCMWorkflowInput _instance;

  final TRes Function(Input$EnqueueSCMWorkflowInput) _then;

  static const _undefined = <dynamic, dynamic>{};

  TRes call({
    Object? operation = _undefined,
    Object? provider = _undefined,
    Object? owner = _undefined,
    Object? repository = _undefined,
    Object? runID = _undefined,
    Object? taskID = _undefined,
    Object? jobID = _undefined,
    Object? idempotencyKey = _undefined,
    Object? worktreePath = _undefined,
    Object? baseBranch = _undefined,
    Object? targetBranch = _undefined,
    Object? pullRequestNumber = _undefined,
    Object? mergeMethod = _undefined,
    Object? pullRequestTitle = _undefined,
    Object? pullRequestBody = _undefined,
    Object? reviewDecision = _undefined,
    Object? reviewBody = _undefined,
  }) => _then(
    Input$EnqueueSCMWorkflowInput._({
      ..._instance._$data,
      if (operation != _undefined && operation != null)
        'operation': (operation as Enum$SCMOperation),
      if (provider != _undefined && provider != null)
        'provider': (provider as Enum$SCMProvider),
      if (owner != _undefined && owner != null) 'owner': (owner as String),
      if (repository != _undefined && repository != null)
        'repository': (repository as String),
      if (runID != _undefined && runID != null) 'runID': (runID as String),
      if (taskID != _undefined && taskID != null) 'taskID': (taskID as String),
      if (jobID != _undefined && jobID != null) 'jobID': (jobID as String),
      if (idempotencyKey != _undefined && idempotencyKey != null)
        'idempotencyKey': (idempotencyKey as String),
      if (worktreePath != _undefined) 'worktreePath': (worktreePath as String?),
      if (baseBranch != _undefined) 'baseBranch': (baseBranch as String?),
      if (targetBranch != _undefined) 'targetBranch': (targetBranch as String?),
      if (pullRequestNumber != _undefined)
        'pullRequestNumber': (pullRequestNumber as int?),
      if (mergeMethod != _undefined)
        'mergeMethod': (mergeMethod as Enum$SCMMergeMethod?),
      if (pullRequestTitle != _undefined)
        'pullRequestTitle': (pullRequestTitle as String?),
      if (pullRequestBody != _undefined)
        'pullRequestBody': (pullRequestBody as String?),
      if (reviewDecision != _undefined)
        'reviewDecision': (reviewDecision as Enum$SCMReviewDecision?),
      if (reviewBody != _undefined) 'reviewBody': (reviewBody as String?),
    }),
  );
}

class _CopyWithStubImpl$Input$EnqueueSCMWorkflowInput<TRes>
    implements CopyWith$Input$EnqueueSCMWorkflowInput<TRes> {
  _CopyWithStubImpl$Input$EnqueueSCMWorkflowInput(this._res);

  TRes _res;

  call({
    Enum$SCMOperation? operation,
    Enum$SCMProvider? provider,
    String? owner,
    String? repository,
    String? runID,
    String? taskID,
    String? jobID,
    String? idempotencyKey,
    String? worktreePath,
    String? baseBranch,
    String? targetBranch,
    int? pullRequestNumber,
    Enum$SCMMergeMethod? mergeMethod,
    String? pullRequestTitle,
    String? pullRequestBody,
    Enum$SCMReviewDecision? reviewDecision,
    String? reviewBody,
  }) => _res;
}

enum Enum$SCMOperation {
  SOURCE_STATE,
  ENSURE_WORKTREE,
  SYNC_WORKTREE,
  CLEANUP_WORKTREE,
  ENSURE_BRANCH,
  SYNC_BRANCH,
  UPSERT_PULL_REQUEST,
  GET_PULL_REQUEST,
  SUBMIT_REVIEW,
  CHECK_MERGE_READINESS,
  MERGE_PULL_REQUEST,
  $unknown;

  factory Enum$SCMOperation.fromJson(String value) =>
      fromJson$Enum$SCMOperation(value);

  String toJson() => toJson$Enum$SCMOperation(this);
}

String toJson$Enum$SCMOperation(Enum$SCMOperation e) {
  switch (e) {
    case Enum$SCMOperation.SOURCE_STATE:
      return r'SOURCE_STATE';
    case Enum$SCMOperation.ENSURE_WORKTREE:
      return r'ENSURE_WORKTREE';
    case Enum$SCMOperation.SYNC_WORKTREE:
      return r'SYNC_WORKTREE';
    case Enum$SCMOperation.CLEANUP_WORKTREE:
      return r'CLEANUP_WORKTREE';
    case Enum$SCMOperation.ENSURE_BRANCH:
      return r'ENSURE_BRANCH';
    case Enum$SCMOperation.SYNC_BRANCH:
      return r'SYNC_BRANCH';
    case Enum$SCMOperation.UPSERT_PULL_REQUEST:
      return r'UPSERT_PULL_REQUEST';
    case Enum$SCMOperation.GET_PULL_REQUEST:
      return r'GET_PULL_REQUEST';
    case Enum$SCMOperation.SUBMIT_REVIEW:
      return r'SUBMIT_REVIEW';
    case Enum$SCMOperation.CHECK_MERGE_READINESS:
      return r'CHECK_MERGE_READINESS';
    case Enum$SCMOperation.MERGE_PULL_REQUEST:
      return r'MERGE_PULL_REQUEST';
    case Enum$SCMOperation.$unknown:
      return r'$unknown';
  }
}

Enum$SCMOperation fromJson$Enum$SCMOperation(String value) {
  switch (value) {
    case r'SOURCE_STATE':
      return Enum$SCMOperation.SOURCE_STATE;
    case r'ENSURE_WORKTREE':
      return Enum$SCMOperation.ENSURE_WORKTREE;
    case r'SYNC_WORKTREE':
      return Enum$SCMOperation.SYNC_WORKTREE;
    case r'CLEANUP_WORKTREE':
      return Enum$SCMOperation.CLEANUP_WORKTREE;
    case r'ENSURE_BRANCH':
      return Enum$SCMOperation.ENSURE_BRANCH;
    case r'SYNC_BRANCH':
      return Enum$SCMOperation.SYNC_BRANCH;
    case r'UPSERT_PULL_REQUEST':
      return Enum$SCMOperation.UPSERT_PULL_REQUEST;
    case r'GET_PULL_REQUEST':
      return Enum$SCMOperation.GET_PULL_REQUEST;
    case r'SUBMIT_REVIEW':
      return Enum$SCMOperation.SUBMIT_REVIEW;
    case r'CHECK_MERGE_READINESS':
      return Enum$SCMOperation.CHECK_MERGE_READINESS;
    case r'MERGE_PULL_REQUEST':
      return Enum$SCMOperation.MERGE_PULL_REQUEST;
    default:
      return Enum$SCMOperation.$unknown;
  }
}

enum Enum$SCMProvider {
  GITHUB,
  $unknown;

  factory Enum$SCMProvider.fromJson(String value) =>
      fromJson$Enum$SCMProvider(value);

  String toJson() => toJson$Enum$SCMProvider(this);
}

String toJson$Enum$SCMProvider(Enum$SCMProvider e) {
  switch (e) {
    case Enum$SCMProvider.GITHUB:
      return r'GITHUB';
    case Enum$SCMProvider.$unknown:
      return r'$unknown';
  }
}

Enum$SCMProvider fromJson$Enum$SCMProvider(String value) {
  switch (value) {
    case r'GITHUB':
      return Enum$SCMProvider.GITHUB;
    default:
      return Enum$SCMProvider.$unknown;
  }
}

enum Enum$SCMMergeMethod {
  MERGE,
  SQUASH,
  REBASE,
  $unknown;

  factory Enum$SCMMergeMethod.fromJson(String value) =>
      fromJson$Enum$SCMMergeMethod(value);

  String toJson() => toJson$Enum$SCMMergeMethod(this);
}

String toJson$Enum$SCMMergeMethod(Enum$SCMMergeMethod e) {
  switch (e) {
    case Enum$SCMMergeMethod.MERGE:
      return r'MERGE';
    case Enum$SCMMergeMethod.SQUASH:
      return r'SQUASH';
    case Enum$SCMMergeMethod.REBASE:
      return r'REBASE';
    case Enum$SCMMergeMethod.$unknown:
      return r'$unknown';
  }
}

Enum$SCMMergeMethod fromJson$Enum$SCMMergeMethod(String value) {
  switch (value) {
    case r'MERGE':
      return Enum$SCMMergeMethod.MERGE;
    case r'SQUASH':
      return Enum$SCMMergeMethod.SQUASH;
    case r'REBASE':
      return Enum$SCMMergeMethod.REBASE;
    default:
      return Enum$SCMMergeMethod.$unknown;
  }
}

enum Enum$SCMReviewDecision {
  APPROVE,
  REQUEST_CHANGES,
  COMMENT,
  $unknown;

  factory Enum$SCMReviewDecision.fromJson(String value) =>
      fromJson$Enum$SCMReviewDecision(value);

  String toJson() => toJson$Enum$SCMReviewDecision(this);
}

String toJson$Enum$SCMReviewDecision(Enum$SCMReviewDecision e) {
  switch (e) {
    case Enum$SCMReviewDecision.APPROVE:
      return r'APPROVE';
    case Enum$SCMReviewDecision.REQUEST_CHANGES:
      return r'REQUEST_CHANGES';
    case Enum$SCMReviewDecision.COMMENT:
      return r'COMMENT';
    case Enum$SCMReviewDecision.$unknown:
      return r'$unknown';
  }
}

Enum$SCMReviewDecision fromJson$Enum$SCMReviewDecision(String value) {
  switch (value) {
    case r'APPROVE':
      return Enum$SCMReviewDecision.APPROVE;
    case r'REQUEST_CHANGES':
      return Enum$SCMReviewDecision.REQUEST_CHANGES;
    case r'COMMENT':
      return Enum$SCMReviewDecision.COMMENT;
    default:
      return Enum$SCMReviewDecision.$unknown;
  }
}
