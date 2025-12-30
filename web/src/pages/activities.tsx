import { useState } from 'react'
import { useActivities } from '@/lib/api'
import { useActivityFiltersStore } from '@/stores'
import {
  ActivitiesTable,
  VirtualizedActivitiesTable,
  Pagination,
  ActivityFiltersPanel,
} from '@/components/activities'
import { Button } from '@/components/ui/button'
import { List, LayoutGrid } from 'lucide-react'
import { PAGINATION } from '@/lib/constants'

export function ActivitiesPage() {
  const { filters, setFilters, resetFilters } = useActivityFiltersStore()
  const [viewAll, setViewAll] = useState(false)

  // When viewAll is true, fetch all activities (up to VIEW_ALL_LIMIT)
  const effectiveFilters = viewAll
    ? { ...filters, per_page: PAGINATION.VIEW_ALL_LIMIT, page: 1 }
    : filters
  const { data, isLoading, error } = useActivities(effectiveFilters)

  const handlePageChange = (page: number) => {
    setFilters({ page })
  }

  const toggleViewAll = () => {
    setViewAll((prev) => !prev)
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold">Activities</h1>
          <p className="text-muted-foreground">Browse and filter your activities</p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={viewAll ? 'default' : 'outline'}
            size="sm"
            onClick={toggleViewAll}
            className="gap-2"
          >
            {viewAll ? (
              <>
                <LayoutGrid className="h-4 w-4" />
                Paginated
              </>
            ) : (
              <>
                <List className="h-4 w-4" />
                View All
              </>
            )}
          </Button>
        </div>
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

      {viewAll ? (
        <VirtualizedActivitiesTable
          activities={data?.data ?? []}
          isLoading={isLoading}
          height={600}
        />
      ) : (
        <>
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
        </>
      )}
    </div>
  )
}
