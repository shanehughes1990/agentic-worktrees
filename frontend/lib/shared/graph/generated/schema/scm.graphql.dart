enum Enum$SCMOperation {
  SOURCE_STATE,
  ENSURE_REPOSITORY,
  SYNC_REPOSITORY,
  CLEANUP_REPOSITORY,
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
    case Enum$SCMOperation.ENSURE_REPOSITORY:
      return r'ENSURE_REPOSITORY';
    case Enum$SCMOperation.SYNC_REPOSITORY:
      return r'SYNC_REPOSITORY';
    case Enum$SCMOperation.CLEANUP_REPOSITORY:
      return r'CLEANUP_REPOSITORY';
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
    case r'ENSURE_REPOSITORY':
      return Enum$SCMOperation.ENSURE_REPOSITORY;
    case r'SYNC_REPOSITORY':
      return Enum$SCMOperation.SYNC_REPOSITORY;
    case r'CLEANUP_REPOSITORY':
      return Enum$SCMOperation.CLEANUP_REPOSITORY;
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
