import { ActivityDetailPage } from '../pages/activity-detail.page'
import { expect, test } from './fixtures'

interface SqlJsQueryResult {
  values: (number | string | Uint8Array | null)[][]
}

interface SqlJsDatabase {
  exec: (sql: string) => SqlJsQueryResult[]
}

test('activity speed preserves the fractional REAL value from the demo database', async ({
  page,
}) => {
  const activity = await page.evaluate(() => {
    const databases = (
      globalThis as unknown as {
        _go_sqlite_dbs: Map<string, SqlJsDatabase>
      }
    )._go_sqlite_dbs
    const database = databases.get(':memory:')
    if (!database) throw new Error('Shared sql.js database was not initialized')

    const [result] = database.exec(`
      SELECT id, average_speed
      FROM activities
      WHERE sport_type IN ('Ride', 'VirtualRide', 'GravelRide', 'MountainBikeRide')
        AND ROUND(average_speed * 3.6, 1) != ROUND(CAST(average_speed AS INTEGER) * 3.6, 1)
        AND ROUND(average_speed * 3.6, 1) != CAST(ROUND(average_speed * 3.6, 1) AS INTEGER)
      ORDER BY id
      LIMIT 1
    `)
    const row = result?.values[0]
    if (!row) throw new Error('Demo database has no suitable fractional cycling speed')

    return {
      id: Number(row[0]),
      averageSpeed: Number(row[1]),
    }
  })

  const activityDetail = new ActivityDetailPage(page)
  await activityDetail.goto(activity.id)
  await activityDetail.waitForActivityLoad()

  const renderedSpeed = await activityDetail.getStatValue('Speed')
  const expectedSpeed = `${(activity.averageSpeed * 3.6).toFixed(1)} km/h`

  expect(expectedSpeed).toMatch(/^\d+\.\d km\/h$/)
  expect(renderedSpeed).toBe(expectedSpeed)
})
