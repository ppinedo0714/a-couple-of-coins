package repository

// Compile-time check: pgxImportJobRepository must satisfy ImportJobRepository.
var _ ImportJobRepository = (*pgxImportJobRepository)(nil)
