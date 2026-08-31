package bootstrap

import (
	"path/filepath"
	"testing"

	"github.com/go-rio/sqlite"
	"github.com/stretchr/testify/require"
)

func TestSessionStore(t *testing.T) {
	db, err := sqlite.Open("file:" + filepath.Join(t.TempDir(), "sess.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	store, err := newSessionStore(db)
	require.NoError(t, err)

	// a missing session is reported via found=false, not an error
	_, found, err := store.Read("missing")
	require.NoError(t, err)
	require.False(t, found)

	require.NoError(t, store.Write("sid", "payload"))
	data, found, err := store.Read("sid")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "payload", data)

	// writing the same id upserts rather than duplicating
	require.NoError(t, store.Write("sid", "updated"))
	data, _, _ = store.Read("sid")
	require.Equal(t, "updated", data)

	// touch reports whether the session existed
	ok, err := store.Touch("sid")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = store.Touch("missing")
	require.NoError(t, err)
	require.False(t, ok)

	require.NoError(t, store.Destroy("sid"))
	_, found, _ = store.Read("sid")
	require.False(t, found)

	// gc drops everything past its lifetime
	require.NoError(t, store.Write("stale", "x"))
	require.NoError(t, store.Gc(0))
	_, found, _ = store.Read("stale")
	require.False(t, found)
}
