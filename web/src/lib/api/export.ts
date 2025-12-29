export interface ExportStats {
  total_activities: number
  first_activity: string | null
  last_activity: string | null
}

export { useExportStats } from '@/lib/data'

export function getExportCSVUrl(params?: {
  after?: string
  before?: string
  sport_type?: string
}): string {
  const searchParams = new URLSearchParams()
  if (params?.after) searchParams.set('after', params.after)
  if (params?.before) searchParams.set('before', params.before)
  if (params?.sport_type) searchParams.set('sport_type', params.sport_type)

  const query = searchParams.toString()
  return `/api/v1/export/activities/csv${query ? `?${query}` : ''}`
}

export function getExportJSONUrl(params?: {
  after?: string
  before?: string
  sport_type?: string
}): string {
  const searchParams = new URLSearchParams()
  if (params?.after) searchParams.set('after', params.after)
  if (params?.before) searchParams.set('before', params.before)
  if (params?.sport_type) searchParams.set('sport_type', params.sport_type)

  const query = searchParams.toString()
  return `/api/v1/export/activities/json${query ? `?${query}` : ''}`
}
