package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/db/lock"
	"github.com/lib/pq"
)

type ResourceTypeNotFoundError struct {
	ID int
}

func (e ResourceTypeNotFoundError) Error() string {
	return fmt.Sprintf("resource type not found: %d", e.ID)
}

//go:generate counterfeiter . ResourceType

type ResourceType interface {
	PipelineRef

	ID() int
	TeamID() int
	TeamName() string
	Name() string
	Type() string
	Privileged() bool
	Source() atc.Source
	Defaults() atc.Source
	Params() atc.Params
	Tags() atc.Tags
	CheckEvery() *atc.CheckEvery
	CheckTimeout() string
	LastCheckStartTime() time.Time
	LastCheckEndTime() time.Time
	CurrentPinnedVersion() atc.Version
	ResourceConfigScopeID() int

	HasWebhook() bool

	SetResourceConfigScope(ResourceConfigScope) error

	CheckPlan(atc.Version, time.Duration, ResourceTypes, atc.Source) atc.CheckPlan
	CreateBuild(context.Context, bool, atc.Plan) (Build, bool, error)

	Version() atc.Version

	Reload() (bool, error)
}

type ResourceTypes []ResourceType

func (resourceTypes ResourceTypes) Parent(checkable Checkable) (ResourceType, bool) {
	for _, t := range resourceTypes {
		if t.PipelineID() == checkable.PipelineID() {
			if t != checkable && t.Name() == checkable.Type() {
				return t, true
			}
		}
	}
	return nil, false
}

func (resourceTypes ResourceTypes) Filter(checkable Checkable) ResourceTypes {
	var result ResourceTypes

	for {
		resourceType, found := resourceTypes.Parent(checkable)
		if !found {
			return result
		}

		result = append(result, resourceType)
		checkable = resourceType
	}
}

func (resourceTypes ResourceTypes) Deserialize() atc.VersionedResourceTypes {
	var versionedResourceTypes atc.VersionedResourceTypes

	for _, t := range resourceTypes {
		// Apply source defaults to resource types
		source := t.Source()
		parentType, found := resourceTypes.Parent(t)
		if found {
			source = parentType.Defaults().Merge(source)
		} else {
			defaults, found := atc.FindBaseResourceTypeDefaults(t.Type())
			if found {
				source = defaults.Merge(source)
			}
		}

		versionedResourceTypes = append(versionedResourceTypes, atc.VersionedResourceType{
			ResourceType: atc.ResourceType{
				Name:       t.Name(),
				Type:       t.Type(),
				Source:     source,
				Defaults:   t.Defaults(),
				Privileged: t.Privileged(),
				CheckEvery: t.CheckEvery(),
				Tags:       t.Tags(),
				Params:     t.Params(),
			},
			Version: t.Version(),
		})
	}

	return versionedResourceTypes
}

func (resourceTypes ResourceTypes) Configs() atc.ResourceTypes {
	var configs atc.ResourceTypes

	for _, r := range resourceTypes {
		configs = append(configs, atc.ResourceType{
			Name:       r.Name(),
			Type:       r.Type(),
			Source:     r.Source(),
			Defaults:   r.Defaults(),
			Privileged: r.Privileged(),
			CheckEvery: r.CheckEvery(),
			Tags:       r.Tags(),
			Params:     r.Params(),
		})
	}

	return configs
}

var resourceTypesQuery = psql.Select(
	"r.id",
	"r.pipeline_id",
	"r.name",
	"r.type",
	"r.config",
	"rcv.version",
	"r.nonce",
	"p.name",
	"p.instance_vars",
	"t.id",
	"t.name",
	"ro.id",
	"ro.last_check_start_time",
	"ro.last_check_end_time",
).
	From("resource_types r").
	Join("pipelines p ON p.id = r.pipeline_id").
	Join("teams t ON t.id = p.team_id").
	LeftJoin("resource_configs c ON c.id = r.resource_config_id").
	LeftJoin("resource_config_scopes ro ON ro.resource_config_id = c.id").
	LeftJoin(`LATERAL (
		SELECT rcv.*
		FROM resource_config_versions rcv
		WHERE rcv.resource_config_scope_id = ro.id
		ORDER BY rcv.check_order DESC
		LIMIT 1
	) AS rcv ON true`).
	Where(sq.Eq{"r.active": true})

type resourceType struct {
	pipelineRef

	id                    int
	teamID                int
	resourceConfigScopeID int
	teamName              string
	name                  string
	type_                 string
	privileged            bool
	source                atc.Source
	defaults              atc.Source
	params                atc.Params
	tags                  atc.Tags
	version               atc.Version
	checkEvery            *atc.CheckEvery
	lastCheckStartTime    time.Time
	lastCheckEndTime      time.Time
}

func (t *resourceType) ID() int                       { return t.id }
func (t *resourceType) TeamID() int                   { return t.teamID }
func (t *resourceType) TeamName() string              { return t.teamName }
func (t *resourceType) Name() string                  { return t.name }
func (t *resourceType) Type() string                  { return t.type_ }
func (t *resourceType) Privileged() bool              { return t.privileged }
func (t *resourceType) CheckEvery() *atc.CheckEvery   { return t.checkEvery }
func (t *resourceType) CheckTimeout() string          { return "" }
func (r *resourceType) LastCheckStartTime() time.Time { return r.lastCheckStartTime }
func (r *resourceType) LastCheckEndTime() time.Time   { return r.lastCheckEndTime }
func (t *resourceType) Source() atc.Source            { return t.source }
func (t *resourceType) Defaults() atc.Source          { return t.defaults }
func (t *resourceType) Params() atc.Params            { return t.params }
func (t *resourceType) Tags() atc.Tags                { return t.tags }
func (t *resourceType) ResourceConfigScopeID() int    { return t.resourceConfigScopeID }

func (t *resourceType) Version() atc.Version              { return t.version }
func (t *resourceType) CurrentPinnedVersion() atc.Version { return nil }

func (t *resourceType) HasWebhook() bool {
	return false
}

func newEmptyResourceType(conn Conn, lockFactory lock.LockFactory) *resourceType {
	return &resourceType{pipelineRef: pipelineRef{conn: conn, lockFactory: lockFactory}}
}

func (t *resourceType) Reload() (bool, error) {
	row := resourceTypesQuery.Where(sq.Eq{"r.id": t.id}).RunWith(t.conn).QueryRow()

	err := scanResourceType(t, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *resourceType) SetResourceConfig(atc.Source, atc.VersionedResourceTypes) (ResourceConfigScope, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *resourceType) SetResourceConfigScope(scope ResourceConfigScope) error {
	_, err := psql.Update("resource_types").
		Set("resource_config_id", scope.ResourceConfig().ID()).
		Where(sq.Eq{"id": r.id}).
		Where(sq.Or{
			sq.Eq{"resource_config_id": nil},
			sq.NotEq{"resource_config_id": scope.ResourceConfig().ID()},
		}).
		RunWith(r.conn).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func (r *resourceType) CheckPlan(from atc.Version, interval time.Duration, resourceTypes ResourceTypes, sourceDefaults atc.Source) atc.CheckPlan {
	return atc.CheckPlan{
		Name:   r.Name(),
		Type:   r.Type(),
		Source: sourceDefaults.Merge(r.Source()),
		Tags:   r.Tags(),

		FromVersion:            from,
		Interval:               interval.String(),
		VersionedResourceTypes: resourceTypes.Deserialize(),

		ResourceType: r.Name(),
	}
}

func (r *resourceType) CreateBuild(ctx context.Context, manuallyTriggered bool, plan atc.Plan) (Build, bool, error) {
	spanContextJSON, err := json.Marshal(NewSpanContext(ctx))
	if err != nil {
		return nil, false, err
	}

	tx, err := r.conn.Begin()
	if err != nil {
		return nil, false, err
	}

	defer Rollback(tx)

	if !manuallyTriggered {
		var buildID int
		var completed, noBuild bool
		err = psql.Select("id", "completed").
			From("builds").
			Where(sq.Eq{"resource_type_id": r.id}).
			RunWith(tx).
			QueryRow().
			Scan(&buildID, &completed)
		if err != nil {
			if err == sql.ErrNoRows {
				noBuild = true
			} else {
				return nil, false, err
			}
		}

		if !noBuild && !completed {
			// a build is already running; leave it be
			return nil, false, nil
		}

		if completed {
			// previous build finished; clear it out
			_, err = psql.Delete("builds").
				Where(sq.Eq{
					"resource_type_id": r.id,
					"completed":        true,
				}).
				RunWith(tx).
				Exec()
			if err != nil {
				return nil, false, fmt.Errorf("delete previous build: %w", err)
			}
			_, err = psql.Delete("build_events").
				Where(sq.Eq{
					"build_id": buildID,
				}).
				RunWith(tx).
				Exec()
			if err != nil {
				return nil, false, fmt.Errorf("delete previous build events: %w", err)
			}
		}
	}

	build := newEmptyBuild(r.conn, r.lockFactory)
	err = createBuild(tx, build, map[string]interface{}{
		"name":               CheckBuildName,
		"team_id":            r.teamID,
		"pipeline_id":        r.pipelineID,
		"resource_type_id":   r.id,
		"status":             BuildStatusPending,
		"manually_triggered": manuallyTriggered,
		"span_context":       string(spanContextJSON),
	})
	if err != nil {
		return nil, false, err
	}

	_, err = build.start(tx, plan)
	if err != nil {
		return nil, false, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, false, err
	}

	err = r.conn.Bus().Notify(atc.ComponentBuildTracker)
	if err != nil {
		return nil, false, err
	}

	_, err = build.Reload()
	if err != nil {
		return nil, false, err
	}

	return build, true, nil
}

func scanResourceType(t *resourceType, row scannable) error {
	var (
		configJSON                           sql.NullString
		rcsID, version, nonce                sql.NullString
		lastCheckStartTime, lastCheckEndTime pq.NullTime
		pipelineInstanceVars                 sql.NullString
	)

	err := row.Scan(&t.id, &t.pipelineID, &t.name, &t.type_, &configJSON, &version, &nonce, &t.pipelineName, &pipelineInstanceVars, &t.teamID, &t.teamName, &rcsID, &lastCheckStartTime, &lastCheckEndTime)
	if err != nil {
		return err
	}

	t.lastCheckStartTime = lastCheckStartTime.Time
	t.lastCheckEndTime = lastCheckEndTime.Time

	if version.Valid {
		err = json.Unmarshal([]byte(version.String), &t.version)
		if err != nil {
			return err
		}
	}

	es := t.conn.EncryptionStrategy()

	var noncense *string
	if nonce.Valid {
		noncense = &nonce.String
	}

	var config atc.ResourceType
	if configJSON.Valid {
		decryptedConfig, err := es.Decrypt(configJSON.String, noncense)
		if err != nil {
			return err
		}

		err = json.Unmarshal(decryptedConfig, &config)
		if err != nil {
			return err
		}
	} else {
		config = atc.ResourceType{}
	}

	t.source = config.Source
	t.defaults = config.Defaults
	t.params = config.Params
	t.privileged = config.Privileged
	t.tags = config.Tags
	t.checkEvery = config.CheckEvery

	if rcsID.Valid {
		t.resourceConfigScopeID, err = strconv.Atoi(rcsID.String)
		if err != nil {
			return err
		}
	}

	if pipelineInstanceVars.Valid {
		err = json.Unmarshal([]byte(pipelineInstanceVars.String), &t.pipelineInstanceVars)
		if err != nil {
			return err
		}
	}

	return nil
}
