package sandbox

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
	"github.com/stretchr/testify/require"
)

func setupSandbox(t *testing.T) (svc *service.Service, cancel func()) {
	username := os.Getenv("UPCLOUD_USERNAME")
	password := os.Getenv("UPCLOUD_PASSWORD")
	if username == "" || password == "" {
		t.Skip("UpCloud credentials not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	sb := New(username, password)
	user, err := sb.Create(ctx)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	return service.New(user), func() {
		if sb == nil {
			return
		}
		if err := sb.Delete(context.Background()); err != nil {
			t.Log(err)
		}
	}
}

// TestSandbox tests sandbox and ucnuke.Account
func TestSandbox(t *testing.T) {
	t.Parallel()

	if os.Getenv("TEST_SANDBOX") == "" {
		t.Skip("env TEST_SANDBOX not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	t.Cleanup(cancel)

	svc, cancelSandbox := setupSandbox(t)
	t.Cleanup(cancelSandbox)

	t.Run("lb_stack", func(t *testing.T) {
		t.Parallel()

		srvIP := "10.0.100.10"
		zone := "pl-waw1"
		router, err := svc.CreateRouter(ctx, createRouterRequest())
		require.NoError(t, err)

		sdn, err := svc.CreateNetwork(ctx, createNetworkRequest(zone, router.UUID))
		require.NoError(t, err)

		_, err = svc.CreateServer(ctx, createServerRequest(zone, srvIP, sdn.UUID))
		require.NoError(t, err)

		_, err = svc.CreateLoadBalancer(ctx, createLoadBalancerRequest(zone, srvIP, sdn.UUID))
		require.NoError(t, err)
	})
	t.Run("storage", func(t *testing.T) {
		t.Parallel()

		_, err := svc.CreateStorage(ctx, createStorageRequest("fi-hel2"))
		require.NoError(t, err)
	})
	t.Run("object storage", func(t *testing.T) {
		t.Parallel()
		_, err := svc.CreateObjectStorage(ctx, createObjectStorageRequest("nl-ams1"))
		require.NoError(t, err)
	})
	t.Run("database", func(t *testing.T) {
		t.Parallel()

		_, err := svc.CreateManagedDatabase(ctx, createManagedDatabaseRequest("de-fra1"))
		require.NoError(t, err)
	})
}

func TestPassword(t *testing.T) {
	t.Parallel()

	passwords := map[string]bool{}
	for i := 0; i <= 100; i++ {
		p := password(10)
		if _, ok := passwords[p]; ok {
			t.Errorf("duplicate password %s generated (%d)", p, i)
		}
		passwords[p] = true
		time.Sleep(time.Millisecond * 2)
	}
}
