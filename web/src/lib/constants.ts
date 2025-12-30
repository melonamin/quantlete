// Pagination constants - keep in sync with internal/pagination/pagination.go
export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_PER_PAGE: 50,
  MAX_PER_PAGE: 200,
  MAX_PAGE: 10000,
  // Maximum items for "view all" mode to prevent performance issues
  VIEW_ALL_LIMIT: 2000,
} as const
