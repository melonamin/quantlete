import {
  useAuthStatus,
  useDeleteHrZoneDefinition,
  useFtpHistory,
  useHrZoneDefinitions,
  useUpdateFtpHistory,
  useUpdateWeightHistory,
  useUpsertHrZoneDefinition,
  useWeightHistory,
} from '@/lib/data/hooks'
import {
  FtpEditor,
  HrZonesEditor,
  WeightEditor,
} from '@/components/settings'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { User, MapPin } from 'lucide-react'

export function AthletePage() {
  const { data: authStatus } = useAuthStatus()
  const isAuthenticated = authStatus?.authenticated
  const athlete = authStatus?.athlete

  const { data: ftpHistory } = useFtpHistory()
  const updateFtp = useUpdateFtpHistory()
  const { data: weightHistory } = useWeightHistory()
  const updateWeight = useUpdateWeightHistory()
  const { data: hrZoneDefs } = useHrZoneDefinitions({ enabled: !!isAuthenticated })
  const upsertHrZones = useUpsertHrZoneDefinition()
  const deleteHrZones = useDeleteHrZoneDefinition()

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-2xl font-bold">Athlete Profile</h1>
          <p className="text-muted-foreground">Your training metrics and zones</p>
        </div>
        <Card className="max-w-2xl">
          <CardContent className="pt-6">
            <p className="text-muted-foreground">
              Connect your Strava account in Settings to view your athlete profile.
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold">Athlete Profile</h1>
        <p className="text-muted-foreground">Your training metrics and zones</p>
      </div>

      <div className="space-y-6 max-w-2xl">
        {/* Athlete Info */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              {athlete?.profile ? (
                <img
                  src={athlete.profile}
                  alt={`${athlete.firstname} ${athlete.lastname}`}
                  className="h-16 w-16 rounded-full object-cover ring-2 ring-border"
                />
              ) : (
                <div className="h-16 w-16 rounded-full bg-muted flex items-center justify-center ring-2 ring-border">
                  <User className="h-8 w-8 text-muted-foreground" />
                </div>
              )}
              <div>
                <h2 className="text-xl font-semibold">
                  {athlete?.firstname} {athlete?.lastname}
                </h2>
                {athlete?.username && (
                  <p className="text-sm text-muted-foreground">@{athlete.username}</p>
                )}
                {(athlete?.city || athlete?.country) && (
                  <p className="text-sm text-muted-foreground flex items-center gap-1 mt-1">
                    <MapPin className="h-3 w-3" />
                    {[athlete.city, athlete.country].filter(Boolean).join(', ')}
                  </p>
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* FTP */}
        <Card>
          <CardHeader>
            <CardTitle>Functional Threshold Power</CardTitle>
            <CardDescription>
              Track your FTP for accurate power zone calculations and training load metrics
            </CardDescription>
          </CardHeader>
          <CardContent>
            <FtpEditor
              value={ftpHistory}
              onSave={(v) => updateFtp.mutate(v)}
              saving={updateFtp.isPending}
            />
          </CardContent>
        </Card>

        {/* Weight */}
        <Card>
          <CardHeader>
            <CardTitle>Weight History</CardTitle>
            <CardDescription>
              Track weight for power-to-weight ratio and normalized metrics
            </CardDescription>
          </CardHeader>
          <CardContent>
            <WeightEditor
              value={weightHistory}
              onSave={(v) => updateWeight.mutate(v)}
              saving={updateWeight.isPending}
            />
          </CardContent>
        </Card>

        {/* HR Zones */}
        <Card>
          <CardHeader>
            <CardTitle>Heart Rate Zones</CardTitle>
            <CardDescription>
              Define zones for training intensity analysis and time-in-zone breakdowns
            </CardDescription>
          </CardHeader>
          <CardContent>
            <HrZonesEditor
              defs={hrZoneDefs}
              onSave={(def) => upsertHrZones.mutate(def)}
              onDelete={(p) => deleteHrZones.mutate(p)}
              saving={upsertHrZones.isPending}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
