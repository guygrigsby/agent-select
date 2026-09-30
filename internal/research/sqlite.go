package research

// The research tool writes SQLite through the pure-Go driver (§2.2 driver
// pin). Research-only exception to the project's Postgres default: the
// artifact is a byte-frozen local export, not a production store.
import (
	_ "modernc.org/sqlite"
)
