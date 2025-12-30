import { useActivities } from '@/lib/api'
import { useActivityFiltersStore } from '@/stores'
import {
  ActivitiesTable,
  Pagination,
  ActivityFiltersPanel,
} from '@/components/activities'

export function ActivitiesPage() {
  const { filters, setFilters, resetFilters } = useActivityFiltersStore()
  const { data, isLoading, error } = useActivities(filters)

  const handlePageChange = (page: number) => {
    setFilters({ page })
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Activities</h1>
        <p className="text-muted-foreground">Browse and filter your activities</p>
      </div>

      <ActivityFiltersPanel
        filters={filters}
        onFiltersChange={setFilters}
        onReset={resetFilters}
      />

      {error && (
        <div className="rounded-lg border border-destructive bg-destructive/10 p-4 mb-6">
          <p className="text-destructive">Failed to load activities: {error.message}</p>
        </div>
      )}

      <ActivitiesTable activities={data?.data ?? []} isLoading={isLoading} />

      {data && data.total > 0 && (
        <Pagination
          page={data.page}
          totalPages={data.total_pages}
          total={data.total}
          perPage={data.per_page}
          onPageChange={handlePageChange}
        />
      )}
    </div>
  )
}
