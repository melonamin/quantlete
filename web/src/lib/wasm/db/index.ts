/**
 * WASM Database module - sql.js with OPFS persistence.
 */

export { WasmDatabase, getDatabase, initializeDatabase } from './sql-js'
export type { WasmDatabaseOptions, QueryResult } from './sql-js'
export { OPFSStorage, IndexedDBStorage, getStorageBackend } from './opfs-storage'
export { migrations, type Migration } from './schema.gen'
