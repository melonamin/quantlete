import { Link } from '@tanstack/react-router'
import { useTrainingLoad } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { LineChart } from '@/components/charts'
import { Button } from '@/components/ui/button'

export function TrainingLoad() {
  const { data, isLoading } = useTrainingLoad()

  const series = [
    {
      name: 'Fitness (CTL)',
      data: (data?.series ?? []).map((p) => ({ x: p.day, y: p.ctl })),
    },
    {
      name: 'Fatigue (ATL)',
      data: (data?.series ?? []).map((p) => ({ x: p.day, y: p.atl })),
    },
    {
      name: 'Form (TSB)',
      data: (data?.series ?? []).map((p) => ({ x: p.day, y: p.tsb })),
      areaStyle: true,
    },
  ]

  return (
    <WidgetWrapper
      title="Training Load"
      action={
        <Button variant="ghost" size="sm" asChild>
          <Link to="/training-load">View details</Link>
        </Button>
      }
      isLoading={isLoading}
    >
      <div className="space-y-3">
        {data?.summary && (
          <div className="grid grid-cols-3 gap-2 text-center">
            <div className="rounded-md border border-border p-2">
              <div className="text-xs text-muted-foreground">CTL</div>
              <div className="font-semibold">{data.summary.ctl.toFixed(1)}</div>
            </div>
            <div className="rounded-md border border-border p-2">
              <div className="text-xs text-muted-foreground">ATL</div>
              <div className="font-semibold">{data.summary.atl.toFixed(1)}</div>
            </div>
            <div className="rounded-md border border-border p-2">
              <div className="text-xs text-muted-foreground">TSB</div>
              <div className="font-semibold">{data.summary.tsb.toFixed(1)}</div>
            </div>
          </div>
        )}
        <LineChart series={series} height={220} showLegend={false} showDataZoom={false} />
      </div>
    </WidgetWrapper>
  )
}
