package pull

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestPullInputsResolve(t *testing.T) {
	t.Run("should not return an error if run from outside a project directory with no to flag is set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "pull_input_test")
		defer teardown()

		var i inputs
		assert.Nil(t, i.Resolve(profile, nil))
	})

	t.Run("when run inside a project directory", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "pull_input_test")
		defer teardown()

		assert.Nil(t, ioutil.WriteFile(
			filepath.Join(profile.WorkingDirectory, local.FileRealmConfig.String()),
			[]byte(`{"config_version":20210101,"app_id":"eggcorn-abcde","name":"eggcorn"}`),
			0666,
		))

		t.Run("should set inputs from app if no flags are set", func(t *testing.T) {
			var i inputs
			assert.Nil(t, i.Resolve(profile, nil))

			assert.Equal(t, profile.WorkingDirectory, i.LocalPath)
			assert.Equal(t, "eggcorn-abcde", i.RemoteApp)
			assert.Equal(t, realm.AppConfigVersion20210101, i.AppVersion)
		})

		t.Run("should return an error if app version flag is different from the project value", func(t *testing.T) {
			i := inputs{AppVersion: realm.AppConfigVersion20200603}
			assert.Equal(t, errConfigVersionMismatch, i.Resolve(profile, nil))
		})
	})

	t.Run("resolving the to flag should work", func(t *testing.T) {
		homeDir, teardown := u.SetupHomeDir("")
		defer teardown()

		for _, tc := range []struct {
			description    string
			targetFlag     string
			expectedTarget string
		}{
			{
				description:    "should expand the to flag to include the user home directory",
				targetFlag:     "~/my/project/root",
				expectedTarget: filepath.Join(homeDir, "my/project/root"),
			},
			{
				description:    "should resolve the to flag to account for relative paths",
				targetFlag:     "../../cmd",
				expectedTarget: filepath.Join(homeDir, "../../cmd"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				profile := mock.NewProfile(t)

				i := inputs{LocalPath: tc.targetFlag}
				assert.Nil(t, i.Resolve(profile, nil))

				assert.Equal(t, tc.expectedTarget, i.LocalPath)
			})
		}
	})
}

func TestPullInputsResolveRemoteApp(t *testing.T) {
	t.Run("should not resolve group id if project is provided", func(t *testing.T) {
		i := inputs{Project: "some-project", RemoteApp: "some-app"}

		var realmClient mock.RealmClient

		var appFilter realm.AppFilter
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			appFilter = filter
			return []realm.App{{GroupID: "group-id", ID: "app-id"}}, nil
		}

		app, err := i.resolveRemoteApp(nil, cli.Clients{Realm: realmClient})
		assert.Nil(t, err)
		assert.Equal(t, realm.App{GroupID: "group-id", ID: "app-id"}, app)

		assert.Equal(t, realm.AppFilter{GroupID: "some-project", App: "some-app"}, appFilter)
	})

	t.Run("should resolve group id if project is not provided", func(t *testing.T) {
		i := inputs{RemoteApp: "some-app"}

		var atlasClient mock.AtlasClient
		atlasClient.GroupsFn = func() ([]atlas.Group, error) {
			return []atlas.Group{{ID: "group-id", Name: "group-name"}}, nil
		}

		var realmClient mock.RealmClient

		var appFilter realm.AppFilter
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			appFilter = filter
			return []realm.App{{GroupID: "group-id", ID: "app-id"}}, nil
		}

		app, err := i.resolveRemoteApp(nil, cli.Clients{Atlas: atlasClient, Realm: realmClient})
		assert.Nil(t, err)
		assert.Equal(t, realm.App{GroupID: "group-id", ID: "app-id"}, app)

		assert.Equal(t, realm.AppFilter{GroupID: "group-id", App: "some-app"}, appFilter)
	})

	t.Run("should return a project not found error if the app is not found", func(t *testing.T) {
		i := inputs{Project: "some-project"}

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, cli.ErrAppNotFound{}
		}

		_, err := i.resolveRemoteApp(nil, cli.Clients{Realm: realmClient})
		assert.Equal(t, errProjectNotFound{}, err)
	})

	t.Run("should return an error when the atlas client fails to find groups", func(t *testing.T) {
		var i inputs

		var atlasClient mock.AtlasClient
		atlasClient.GroupsFn = func() ([]atlas.Group, error) {
			return nil, errors.New("something bad happened")
		}

		_, err := i.resolveRemoteApp(nil, cli.Clients{Atlas: atlasClient})
		assert.Equal(t, errors.New("something bad happened"), err)
	})

	t.Run("should return an error when the realm client fails to find an app", func(t *testing.T) {
		i := inputs{Project: "some-project"}

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("something bad happened")
		}

		_, err := i.resolveRemoteApp(nil, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)
	})
}
