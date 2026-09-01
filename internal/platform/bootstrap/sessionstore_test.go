package bootstrap

import (
	"path/filepath"
	"testing"

	"github.com/go-rio/sqlite"
	"github.com/libtnb/assert/must"
)

func TestSessionStore(t *testing.T) {
	db, err := sqlite.Open("file:" + filepath.Join(t.TempDir(), "sess.db"))
	must.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	store, err := newSessionStore(db)
	must.NoError(t, err)

	// a missing session is reported via found=false, not an error
	_, found, err := store.Read("missing")
	must.NoError(t, err)
	must.False(t, found)

	must.NoError(t, store.Write("sid", "payload"))
	data, found, err := store.Read("sid")
	must.NoError(t, err)
	must.True(t, found)
	must.Equal(t, data, "payload")

	// writing the same id upserts rather than duplicating
	must.NoError(t, store.Write("sid", "updated"))
	data, _, _ = store.Read("sid")
	must.Equal(t, data, "updated")

	// touch reports whether the session existed
	ok, err := store.Touch("sid")
	must.NoError(t, err)
	must.True(t, ok)
	ok, err = store.Touch("missing")
	must.NoError(t, err)
	must.False(t, ok)

	must.NoError(t, store.Destroy("sid"))
	_, found, _ = store.Read("sid")
	must.False(t, found)

	// gc drops everything past its lifetime
	must.NoError(t, store.Write("stale", "x"))
	must.NoError(t, store.Gc(0))
	_, found, _ = store.Read("stale")
	must.False(t, found)
}
