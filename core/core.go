package core

import (
	"github.com/GeertJohan/go.rice"
	"github.com/hunterlong/statup/notifiers"
	"github.com/hunterlong/statup/plugin"
	"github.com/hunterlong/statup/types"
	"github.com/pkg/errors"
	"os"
	"time"
)

type PluginJSON types.PluginJSON
type PluginRepos types.PluginRepos

type Core struct {
	Name           string     `db:"name" json:"name"`
	Description    string     `db:"description" json:"name"`
	Config         string     `db:"config" json:"-"`
	ApiKey         string     `db:"api_key" json:"-"`
	ApiSecret      string     `db:"api_secret" json:"-"`
	Style          string     `db:"style" json:"-"`
	Footer         string     `db:"footer" json:"-"`
	Domain         string     `db:"domain" json:"domain,omitempty"`
	Version        string     `db:"version" json:"version,omitempty"`
	MigrationId    int64      `db:"migration_id" json:"-"`
	UseCdn         bool       `db:"use_cdn" json:"-"`
	Services       []*Service `json:"services,omitempty"`
	Plugins        []plugin.Info
	Repos          []PluginJSON
	AllPlugins     []plugin.PluginActions
	Communications []notifiers.AllNotifiers
	DbConnection   string
	started        time.Time
}

var (
	Configs     *types.Config
	CoreApp     *Core
	SqlBox      *rice.Box
	CssBox      *rice.Box
	ScssBox     *rice.Box
	JsBox       *rice.Box
	TmplBox     *rice.Box
	SetupMode   bool
	UsingAssets bool
	VERSION     string
)

func init() {
	CoreApp = NewCore()
}

func NewCore() *Core {
	CoreApp = new(Core)
	CoreApp.started = time.Now()
	return CoreApp
}

func (c *Core) Insert() error {
	col := DbSession.Collection("core")
	_, err := col.Insert(c)
	return err
}

func InitApp() {
	SelectCore()
	notifiers.Collections = DbSession.Collection("communication")
	SelectAllServices()
	CheckServices()
	CoreApp.Communications = notifiers.Load()
	go DatabaseMaintence()
}

func (c *Core) Update() (*Core, error) {
	res := DbSession.Collection("core").Find().Limit(1)
	err := res.Update(c)
	return c, err
}

func (c Core) UsingAssets() bool {
	return UsingAssets
}

func (c Core) SassVars() string {
	if !UsingAssets {
		return ""
	}
	return OpenAsset("scss/variables.scss")
}

func (c Core) BaseSASS() string {
	if !UsingAssets {
		return ""
	}
	return OpenAsset("scss/base.scss")
}

func (c Core) MobileSASS() string {
	if !UsingAssets {
		return ""
	}
	return OpenAsset("scss/mobile.scss")
}

func (c Core) AllOnline() bool {
	for _, ser := range CoreApp.Services {
		s := ser.ToService()
		if !s.Online {
			return false
		}
	}
	return true
}

func SelectLastMigration() (int64, error) {
	var c *Core
	if DbSession == nil {
		return 0, errors.New("Database connection has not been created yet")
	}
	err := DbSession.Collection("core").Find().One(&c)
	if err != nil {
		return 0, err
	}
	return c.MigrationId, err
}

func SelectCore() (*Core, error) {
	var c *Core
	exists := DbSession.Collection("core").Exists()
	if !exists {
		return nil, errors.New("core database has not been setup yet.")
	}

	err := DbSession.Collection("core").Find().One(&c)
	if err != nil {
		return nil, err
	}
	CoreApp = c
	CoreApp.DbConnection = Configs.Connection
	CoreApp.Version = VERSION
	CoreApp.Services, _ = SelectAllServices()
	if os.Getenv("USE_CDN") == "true" {
		CoreApp.UseCdn = true
	}
	//store = sessions.NewCookieStore([]byte(core.ApiSecret))
	return CoreApp, err
}
