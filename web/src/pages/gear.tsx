import { useMemo, useState } from 'react'
import {
  useCreateComponent,
  useCreateCustomGear,
  useDeleteComponent,
  useDeleteCustomGear,
  useGear,
  useGearComponents,
  useGearMonthlyUsage,
  useLogMaintenance,
  useMaintenanceDue,
  useUpdateComponent,
  useUpdateCustomGear,
  type ComponentWithRules,
  type CreateComponentRequest,
  type Gear,
  type GearMonthlyUsage,
  type UpdateComponentRequest,
} from '@/lib/api'
import { formatDistance } from '@/lib/format'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { DonutChart, LineChart, StackedBarChart } from '@/components/charts'
import { Bike, Footprints, Check, Wrench } from 'lucide-react'
import type { ApiError } from '@/lib/api/client'

export function GearPage() {
  const [includeRetired, setIncludeRetired] = useState(false)
  const [tab, setTab] = useState<'gear' | 'maintenance'>('gear')
  const { data: gear, isLoading, error } = useGear({ include_retired: includeRetired })
  const { data: monthlyUsage } = useGearMonthlyUsage(includeRetired)

  const [customModal, setCustomModal] = useState<{ mode: 'create' | 'edit'; gear?: Gear } | null>(
    null
  )
  const createCustom = useCreateCustomGear()
  const updateCustom = useUpdateCustomGear()
  const deleteCustom = useDeleteCustomGear()

  const totalsByGear = useMemo(() => {
    const totals = new Map<string, { movingTime: number; distance: number }>()
    for (const row of monthlyUsage ?? []) {
      const prev = totals.get(row.gear_id) ?? { movingTime: 0, distance: 0 }
      prev.movingTime += row.moving_time || 0
      prev.distance += row.distance || 0
      totals.set(row.gear_id, prev)
    }
    return totals
  }, [monthlyUsage])

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Gear</h1>
          <p className="text-muted-foreground">Manage your equipment and track usage</p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={tab === 'gear' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setTab('gear')}
          >
            Gear
          </Button>
          <Button
            variant={tab === 'maintenance' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setTab('maintenance')}
          >
            Maintenance
          </Button>
          <Button variant="outline" size="sm" onClick={() => setIncludeRetired(!includeRetired)}>
            {includeRetired ? 'Hide Retired' : 'Show Retired'}
          </Button>
          {tab === 'gear' && (
            <Button size="sm" onClick={() => setCustomModal({ mode: 'create' })}>
              Add Custom Gear
            </Button>
          )}
        </div>
      </div>

      {tab === 'gear' ? (
        <>
          {monthlyUsage && monthlyUsage.length > 0 && (
            <GearCharts gear={gear ?? []} usage={monthlyUsage} />
          )}

          {isLoading ? (
            <GearSkeleton />
          ) : error ? (
            <div className="rounded-lg border border-destructive bg-destructive/10 p-8 text-center">
              <p className="text-destructive">Failed to load gear</p>
            </div>
          ) : gear && gear.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {gear.map((item) => (
                <GearCard
                  key={item.id}
                  gear={item}
                  totals={totalsByGear.get(item.id)}
                  onEdit={
                    item.source === 'custom'
                      ? () => setCustomModal({ mode: 'edit', gear: item })
                      : undefined
                  }
                />
              ))}
            </div>
          ) : (
            <div className="rounded-lg border border-border bg-card p-8 text-center">
              <p className="text-muted-foreground">
                No gear found. Import activities with gear to see them here.
              </p>
            </div>
          )}

          {customModal && (
            <CustomGearModal
              modal={customModal}
              onClose={() => setCustomModal(null)}
              onCreate={(req) =>
                createCustom.mutate(req, { onSuccess: () => setCustomModal(null) })
              }
              onUpdate={(id, patch) =>
                updateCustom.mutate({ id, patch }, { onSuccess: () => setCustomModal(null) })
              }
              onDelete={(id, force) =>
                deleteCustom.mutate({ id, force }, { onSuccess: () => setCustomModal(null) })
              }
              createError={createCustom.error as ApiError | null}
              updateError={updateCustom.error as ApiError | null}
              deleteError={deleteCustom.error as ApiError | null}
              isPending={createCustom.isPending || updateCustom.isPending || deleteCustom.isPending}
            />
          )}
        </>
      ) : (
        <MaintenancePanel gear={gear ?? []} />
      )}
    </div>
  )
}

function formatMoney(amount: number, currency?: string) {
  if (!Number.isFinite(amount)) return '–'
  const c = (currency || '').trim().toUpperCase()
  const prefix = c ? `${c} ` : ''
  return `${prefix}${amount.toFixed(2)}`
}

function GearCard({
  gear,
  totals,
  onEdit,
}: {
  gear: Gear
  totals?: { movingTime: number; distance: number }
  onEdit?: () => void
}) {
  const isBike = gear.id.startsWith('b') || gear.source === 'strava'
  const isCustom = gear.source === 'custom'
  const totalHours = totals ? totals.movingTime / 3600 : 0
  const costPerHour =
    gear.purchase_price && totalHours > 0 ? gear.purchase_price / totalHours : null
  const costPerActivity =
    gear.purchase_price && gear.activity_count > 0
      ? gear.purchase_price / gear.activity_count
      : null

  return (
    <Card className={gear.retired ? 'opacity-60' : ''}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            {isCustom ? (
              <Wrench className="h-5 w-5 text-sport-other" />
            ) : isBike ? (
              <Bike className="h-5 w-5 text-sport-ride" />
            ) : (
              <Footprints className="h-5 w-5 text-sport-run" />
            )}
            <CardTitle className="text-lg">{gear.name}</CardTitle>
          </div>
          <div className="flex gap-1">
            {gear.primary && (
              <Badge variant="secondary" className="gap-1">
                <Check className="h-3 w-3" />
                Primary
              </Badge>
            )}
            {isCustom && <Badge variant="outline">Custom</Badge>}
            {gear.retired && <Badge variant="outline">Retired</Badge>}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {(gear.brand_name || gear.model_name) && !isCustom && (
            <p className="text-sm text-muted-foreground">
              {[gear.brand_name, gear.model_name].filter(Boolean).join(' ')}
            </p>
          )}
          {gear.hashtag && <div className="text-sm text-muted-foreground">{gear.hashtag}</div>}
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Distance</span>
            <span className="font-medium">{formatDistance(gear.distance)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Activities</span>
            <span className="font-medium">{gear.activity_count}</span>
          </div>
          {gear.purchase_price != null && (
            <>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Purchase</span>
                <span className="font-medium">
                  {formatMoney(gear.purchase_price, gear.purchase_currency)}
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Cost / hour</span>
                <span className="font-medium">
                  {costPerHour != null ? formatMoney(costPerHour, gear.purchase_currency) : '–'}
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Cost / activity</span>
                <span className="font-medium">
                  {costPerActivity != null
                    ? formatMoney(costPerActivity, gear.purchase_currency)
                    : '–'}
                </span>
              </div>
            </>
          )}
          {onEdit && (
            <div className="pt-2">
              <Button variant="outline" size="sm" onClick={onEdit}>
                Edit
              </Button>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

function GearSkeleton() {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 6 }).map((_, i) => (
        <Card key={i}>
          <CardHeader className="pb-3">
            <div className="flex items-center gap-2">
              <Skeleton className="h-5 w-5" />
              <Skeleton className="h-5 w-32" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-full" />
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

function GearCharts({ gear, usage }: { gear: Gear[]; usage: GearMonthlyUsage[] }) {
  const { months, seriesByGear, cumulativeSeries, movingTimeSlices } = useMemo(() => {
    const months = Array.from(new Set(usage.map((r) => r.month))).sort()
    const gearMeta = new Map<string, { name: string }>()
    for (const row of usage) {
      const name = row.gear_name || row.gear_id
      gearMeta.set(row.gear_id, { name })
    }
    for (const g of gear) {
      if (!gearMeta.has(g.id)) gearMeta.set(g.id, { name: g.name })
    }

    const distanceByGear = new Map<string, number[]>()
    const movingByGear = new Map<string, number[]>()

    for (const id of gearMeta.keys()) {
      distanceByGear.set(
        id,
        Array.from({ length: months.length }, () => 0)
      )
      movingByGear.set(
        id,
        Array.from({ length: months.length }, () => 0)
      )
    }

    for (const row of usage) {
      const idx = months.indexOf(row.month)
      if (idx < 0) continue
      distanceByGear.get(row.gear_id)?.splice(idx, 1, row.distance / 1000)
      movingByGear.get(row.gear_id)?.splice(idx, 1, row.moving_time / 3600)
    }

    const totals = Array.from(distanceByGear.entries()).map(([id, dist]) => ({
      id,
      total: dist.reduce((a, b) => a + b, 0),
      name: gearMeta.get(id)?.name ?? id,
    }))
    totals.sort((a, b) => b.total - a.total)
    const top = totals.slice(0, 8)

    const seriesByGear = top.map((g) => ({
      name: g.name,
      data: distanceByGear.get(g.id) ?? [],
    }))

    const cumulativeSeries = top.map((g) => {
      const d = distanceByGear.get(g.id) ?? []
      let acc = 0
      return {
        name: g.name,
        data: d.map((x, i) => {
          acc += x
          return { x: months[i], y: acc }
        }),
      }
    })

    const movingTimeSlices = top
      .map((g) => ({
        name: g.name,
        value: Math.round((movingByGear.get(g.id) ?? []).reduce((a, b) => a + b, 0) * 10) / 10,
      }))
      .filter((d) => d.value > 0)

    return { months, seriesByGear, cumulativeSeries, movingTimeSlices }
  }, [gear, usage])

  return (
    <div className="mb-6 grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Distance per month (top gear)</CardTitle>
        </CardHeader>
        <CardContent>
          <StackedBarChart categories={months} series={seriesByGear} height={280} />
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Cumulative distance (top gear)</CardTitle>
        </CardHeader>
        <CardContent>
          <LineChart
            xAxisType="category"
            series={cumulativeSeries}
            yAxisLabel="km"
            height={280}
            showLegend={true}
            showDataZoom={true}
          />
        </CardContent>
      </Card>
      <Card className="lg:col-span-2">
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Moving time share (hours)</CardTitle>
        </CardHeader>
        <CardContent>
          <DonutChart data={movingTimeSlices} height={280} />
        </CardContent>
      </Card>
    </div>
  )
}

function CustomGearModal({
  modal,
  onClose,
  onCreate,
  onUpdate,
  onDelete,
  createError,
  updateError,
  deleteError,
  isPending,
}: {
  modal: { mode: 'create' | 'edit'; gear?: Gear }
  onClose: () => void
  onCreate: (req: {
    name: string
    hashtag: string
    retired?: boolean
    purchase_price?: number | null
    purchase_currency?: string
  }) => void
  onUpdate: (id: string, patch: Record<string, unknown>) => void
  onDelete: (id: string, force?: boolean) => void
  createError: ApiError | null
  updateError: ApiError | null
  deleteError: ApiError | null
  isPending: boolean
}) {
  const editing = modal.mode === 'edit' ? modal.gear : undefined
  const [name, setName] = useState(editing?.name ?? '')
  const [hashtag, setHashtag] = useState(editing?.hashtag ?? '')
  const [retired, setRetired] = useState(editing?.retired ?? false)
  const [purchasePrice, setPurchasePrice] = useState(
    editing?.purchase_price != null ? String(editing.purchase_price) : ''
  )
  const [purchaseCurrency, setPurchaseCurrency] = useState(editing?.purchase_currency ?? '')
  const [forceDelete, setForceDelete] = useState(false)

  const error = createError || updateError || deleteError

  return (
    <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
      <div className="mx-auto mt-16 w-[calc(100%-2rem)] max-w-lg rounded-lg border border-border bg-background shadow-lg">
        <div className="flex items-center justify-between border-b border-border px-4 py-3">
          <div className="font-semibold">{editing ? 'Edit Custom Gear' : 'Add Custom Gear'}</div>
          <Button variant="outline" size="sm" onClick={onClose} disabled={isPending}>
            Close
          </Button>
        </div>
        <div className="space-y-3 p-4">
          {error && (
            <div className="rounded-md border border-destructive bg-destructive/10 p-3 text-sm text-destructive">
              {error.message}
            </div>
          )}
          <div>
            <div className="mb-1 text-sm text-muted-foreground">Name</div>
            <input
              className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={isPending}
            />
          </div>
          <div>
            <div className="mb-1 text-sm text-muted-foreground">Hashtag</div>
            <input
              className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
              value={hashtag}
              onChange={(e) => setHashtag(e.target.value)}
              placeholder="#skateboard"
              disabled={isPending}
            />
            <div className="mt-1 text-xs text-muted-foreground">
              Use this tag in activity titles to link the activity to this gear (only when Strava
              gear is empty).
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div>
              <div className="mb-1 text-sm text-muted-foreground">Purchase price</div>
              <input
                className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
                value={purchasePrice}
                onChange={(e) => setPurchasePrice(e.target.value)}
                placeholder="e.g. 499.99"
                disabled={isPending}
              />
            </div>
            <div>
              <div className="mb-1 text-sm text-muted-foreground">Currency</div>
              <input
                className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
                value={purchaseCurrency}
                onChange={(e) => setPurchaseCurrency(e.target.value)}
                placeholder="USD"
                disabled={isPending}
              />
            </div>
          </div>

          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={retired}
              onChange={(e) => setRetired(e.target.checked)}
              disabled={isPending}
            />
            Retired
          </label>

          <div className="flex items-center justify-between gap-2 pt-2">
            {editing ? (
              <>
                <div className="flex items-center gap-2">
                  <Button
                    onClick={() => {
                      const price = purchasePrice.trim() === '' ? null : Number(purchasePrice)
                      onUpdate(editing.id, {
                        name,
                        hashtag,
                        retired,
                        purchase_price: Number.isFinite(price as number) ? price : null,
                        purchase_currency: purchaseCurrency,
                      })
                    }}
                    disabled={isPending}
                  >
                    Save
                  </Button>
                  <Button
                    variant="destructive"
                    onClick={() => onDelete(editing.id, forceDelete)}
                    disabled={isPending}
                  >
                    Delete
                  </Button>
                </div>
                <label className="flex items-center gap-2 text-xs text-muted-foreground">
                  <input
                    type="checkbox"
                    checked={forceDelete}
                    onChange={(e) => setForceDelete(e.target.checked)}
                  />
                  Force delete (unlink activities)
                </label>
              </>
            ) : (
              <Button
                onClick={() => {
                  const price = purchasePrice.trim() === '' ? undefined : Number(purchasePrice)
                  onCreate({
                    name,
                    hashtag,
                    retired,
                    purchase_price: Number.isFinite(price as number)
                      ? (price as number)
                      : undefined,
                    purchase_currency: purchaseCurrency || undefined,
                  })
                }}
                disabled={isPending}
              >
                Create
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function MaintenancePanel({ gear }: { gear: Gear[] }) {
  const { data: due, isLoading: dueLoading } = useMaintenanceDue()
  const gearByID = useMemo(() => new Map(gear.map((g) => [g.id, g])), [gear])

  const [selectedGearId, setSelectedGearId] = useState(() => gear[0]?.id ?? '')
  const { data: components } = useGearComponents(selectedGearId)

  const createComponent = useCreateComponent()
  const updateComponent = useUpdateComponent()
  const deleteComponent = useDeleteComponent()
  const logMaintenance = useLogMaintenance()

  const [componentModal, setComponentModal] = useState<{
    mode: 'create' | 'edit'
    component?: ComponentWithRules
  } | null>(null)

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Maintenance due</CardTitle>
        </CardHeader>
        <CardContent>
          {dueLoading ? (
            <Skeleton className="h-24 w-full" />
          ) : (due ?? []).length === 0 ? (
            <div className="text-sm text-muted-foreground">No components configured yet.</div>
          ) : (
            <div className="space-y-3">
              {(due ?? []).map((c) => (
                <div key={c.id} className="rounded-md border border-border p-3">
                  <div className="flex items-start justify-between gap-4">
                    <div className="min-w-0">
                      <div className="font-medium truncate">{c.name}</div>
                      <div className="text-xs text-muted-foreground">
                        {gearByID.get(c.gear_id)?.name ?? c.gear_id}
                        {c.maintenance_hashtag
                          ? ` • #${c.maintenance_hashtag.replace(/^#/, '')}`
                          : ''}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {c.is_due ? (
                        <Badge variant="destructive">Due</Badge>
                      ) : (
                        <Badge variant="secondary">OK</Badge>
                      )}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => logMaintenance.mutate({ componentId: c.id })}
                        disabled={logMaintenance.isPending}
                      >
                        Log
                      </Button>
                    </div>
                  </div>
                  <div className="mt-3 space-y-2">
                    {c.progress.map((p) => (
                      <div key={p.type}>
                        <div className="flex items-center justify-between text-xs text-muted-foreground">
                          <span>{p.type}</span>
                          <span>{Math.round(p.percent)}%</span>
                        </div>
                        <div className="h-2 w-full rounded bg-muted">
                          <div
                            className={`h-2 rounded ${p.due ? 'bg-destructive' : 'bg-primary'}`}
                            style={{ width: `${Math.min(100, p.percent)}%` }}
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
          <div className="mt-3 text-xs text-muted-foreground">
            Tip: add a component hashtag (e.g. <span className="font-mono">#chain</span>) to an
            activity title to auto-log maintenance on import.
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Manage components</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-2">
              <div className="text-sm text-muted-foreground">Gear</div>
              <select
                className="h-9 rounded-md border border-border bg-background px-2 text-sm"
                value={selectedGearId}
                onChange={(e) => setSelectedGearId(e.target.value)}
              >
                <option value="">Select gear</option>
                {gear.map((g) => (
                  <option key={g.id} value={g.id}>
                    {g.name}
                  </option>
                ))}
              </select>
            </div>
            <Button
              size="sm"
              onClick={() => setComponentModal({ mode: 'create' })}
              disabled={!selectedGearId}
            >
              Add component
            </Button>
          </div>

          {!selectedGearId ? (
            <div className="text-sm text-muted-foreground">
              Select a gear item to configure components.
            </div>
          ) : (components ?? []).length === 0 ? (
            <div className="text-sm text-muted-foreground">No components yet.</div>
          ) : (
            <div className="space-y-2">
              {(components ?? []).map((c) => (
                <div key={c.id} className="rounded-md border border-border p-3">
                  <div className="flex items-start justify-between gap-4">
                    <div className="min-w-0">
                      <div className="font-medium truncate">{c.name}</div>
                      <div className="text-xs text-muted-foreground">
                        {c.maintenance_hashtag
                          ? `#${c.maintenance_hashtag.replace(/^#/, '')}`
                          : 'No hashtag'}
                        {c.last_completed_at
                          ? ` • last: ${new Date(c.last_completed_at).toLocaleDateString()}`
                          : ''}
                      </div>
                      <div className="mt-2 text-xs text-muted-foreground">
                        {c.rules.length === 0
                          ? 'No rules'
                          : c.rules.map((r) => `${r.type}: ${r.threshold_value}`).join(' • ')}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setComponentModal({ mode: 'edit', component: c })}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => logMaintenance.mutate({ componentId: c.id })}
                        disabled={logMaintenance.isPending}
                      >
                        Log
                      </Button>
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => deleteComponent.mutate({ id: c.id })}
                        disabled={deleteComponent.isPending}
                      >
                        Delete
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {componentModal && (
        <ComponentModal
          gearId={selectedGearId}
          modal={componentModal}
          onClose={() => setComponentModal(null)}
          onCreate={(req) =>
            createComponent.mutate(
              { gearId: selectedGearId, body: req },
                { onSuccess: () => setComponentModal(null) }
              )
          }
          onUpdate={(id, req) =>
            updateComponent.mutate({ id, req }, { onSuccess: () => setComponentModal(null) })
          }
        />
      )}
    </div>
  )
}

function ComponentModal({
  gearId,
  modal,
  onClose,
  onCreate,
  onUpdate,
}: {
  gearId: string
  modal: { mode: 'create' | 'edit'; component?: ComponentWithRules }
  onClose: () => void
  onCreate: (req: CreateComponentRequest) => void
  onUpdate: (id: number, req: UpdateComponentRequest) => void
}) {
  const editing = modal.mode === 'edit' ? modal.component : undefined
  const [name, setName] = useState(editing?.name ?? '')
  const [maintenanceHashtag, setMaintenanceHashtag] = useState(editing?.maintenance_hashtag ?? '')
  const [rules, setRules] = useState(() =>
    (editing?.rules ?? []).map((r) => ({
      type: r.type,
      threshold_value: String(r.threshold_value),
    }))
  )

  const addRule = () => {
    setRules((prev) => [...prev, { type: 'distance_m' as const, threshold_value: '' }])
  }

  const parsedRules = () => {
    return rules
      .map((r) => ({
        type: r.type,
        threshold_value: Number(r.threshold_value),
      }))
      .filter((r) => r.type && Number.isFinite(r.threshold_value) && r.threshold_value > 0)
  }

  return (
    <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
      <div className="mx-auto mt-16 w-[calc(100%-2rem)] max-w-lg rounded-lg border border-border bg-background shadow-lg">
        <div className="flex items-center justify-between border-b border-border px-4 py-3">
          <div className="font-semibold">{editing ? 'Edit component' : 'Add component'}</div>
          <Button variant="outline" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>
        <div className="space-y-3 p-4">
          <div className="text-xs text-muted-foreground">Gear: {gearId}</div>
          <div>
            <div className="mb-1 text-sm text-muted-foreground">Name</div>
            <input
              className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div>
            <div className="mb-1 text-sm text-muted-foreground">Maintenance hashtag</div>
            <input
              className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
              value={maintenanceHashtag}
              onChange={(e) => setMaintenanceHashtag(e.target.value)}
              placeholder="#chain"
            />
          </div>

          <div className="rounded-md border border-border p-3 space-y-2">
            <div className="flex items-center justify-between">
              <div className="text-sm font-medium">Rules</div>
              <Button variant="outline" size="sm" onClick={addRule}>
                Add rule
              </Button>
            </div>
            {rules.length === 0 ? (
              <div className="text-sm text-muted-foreground">No rules yet.</div>
            ) : (
              <div className="space-y-2">
                {rules.map((r, idx) => (
                  <div key={idx} className="grid grid-cols-3 gap-2 items-center">
                    <select
                      className="h-9 rounded-md border border-border bg-background px-2 text-sm"
                      value={r.type}
                      onChange={(e) => {
                        const v = e.target.value as 'distance_m' | 'time_s' | 'days'
                        setRules((prev) => prev.map((x, i) => (i === idx ? { ...x, type: v } : x)))
                      }}
                    >
                      <option value="distance_m">distance_m</option>
                      <option value="time_s">time_s</option>
                      <option value="days">days</option>
                    </select>
                    <input
                      className="col-span-2 h-9 rounded-md border border-border bg-background px-2 text-sm"
                      value={r.threshold_value}
                      onChange={(e) =>
                        setRules((prev) =>
                          prev.map((x, i) =>
                            i === idx ? { ...x, threshold_value: e.target.value } : x
                          )
                        )
                      }
                      placeholder="Threshold value"
                    />
                  </div>
                ))}
              </div>
            )}
            <div className="text-xs text-muted-foreground">
              Distance is meters, time is seconds, days is a day count. Example:{' '}
              <span className="font-mono">distance_m: 100000</span> (100 km).
            </div>
          </div>

          <div className="pt-2">
            {editing ? (
              <Button
                onClick={() =>
                  onUpdate(editing.id, {
                    name,
                    maintenance_hashtag: maintenanceHashtag,
                    rules: parsedRules(),
                  })
                }
              >
                Save
              </Button>
            ) : (
              <Button
                onClick={() =>
                  onCreate({ name, maintenance_hashtag: maintenanceHashtag, rules: parsedRules() })
                }
              >
                Create
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
