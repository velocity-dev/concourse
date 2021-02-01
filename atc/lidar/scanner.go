package lidar

import (
	"context"
	"strconv"
	"sync"

	"code.cloudfoundry.org/lager/lagerctx"
	"github.com/concourse/concourse/atc/db"
	"github.com/concourse/concourse/atc/metric"
	"github.com/concourse/concourse/atc/util"
	"github.com/concourse/concourse/tracing"
)

func NewScanner(checkFactory db.CheckFactory) *scanner {
	return &scanner{
		checkFactory: checkFactory,
	}
}

type scanner struct {
	checkFactory db.CheckFactory
}

func (s *scanner) Run(ctx context.Context) error {
	logger := lagerctx.FromContext(ctx)

	spanCtx, span := tracing.StartSpan(ctx, "scanner.Run", nil)
	defer span.End()

	logger.Info("start")
	defer logger.Info("end")

	resources, err := s.checkFactory.Resources()
	if err != nil {
		logger.Error("failed-to-get-resources", err)
		return err
	}

	resourceTypes, err := s.checkFactory.ResourceTypes()
	if err != nil {
		logger.Error("failed-to-get-resource-types", err)
		return err
	}

	waitGroup := new(sync.WaitGroup)
	resourceTypesChecked := &sync.Map{}

	for _, resource := range resources {
		waitGroup.Add(1)

		go func(resource db.Resource, resourceTypes db.ResourceTypes) {
			defer func() {
				err := util.DumpPanic(recover(), "scanning resource %d", resource.ID())
				if err != nil {
					logger.Error("panic-in-scanner-run", err)
				}
			}()
			defer waitGroup.Done()

			if resource.CheckEvery() != nil && resource.CheckEvery().Never {
				return
			}

			s.check(spanCtx, resource, resourceTypes, resourceTypesChecked)
		}(resource, resourceTypes)
	}

	waitGroup.Wait()

	return nil
}

func (s *scanner) check(ctx context.Context, checkable db.Checkable, resourceTypes db.ResourceTypes, resourceTypesChecked *sync.Map) {
	logger := lagerctx.FromContext(ctx)

	spanCtx, span := tracing.StartSpan(ctx, "scanner.check", tracing.Attrs{
		"team":                     checkable.TeamName(),
		"pipeline":                 checkable.PipelineName(),
		"resource":                 checkable.Name(),
		"type":                     checkable.Type(),
		"resource_config_scope_id": strconv.Itoa(checkable.ResourceConfigScopeID()),
	})
	defer span.End()

	parentType, found := resourceTypes.Parent(checkable)
	if found {
		if _, exists := resourceTypesChecked.LoadOrStore(parentType.ID(), true); !exists {
			// only create a check for resource type if it has not been checked yet
			s.check(spanCtx, parentType, resourceTypes, resourceTypesChecked)
		}
	}

	version := checkable.CurrentPinnedVersion()

	_, created, err := s.checkFactory.TryCreateCheck(lagerctx.NewContext(spanCtx, logger), checkable, resourceTypes, version, false)
	if err != nil {
		logger.Error("failed-to-create-check", err)
		return
	}

	if !created {
		logger.Debug("check-already-exists")
	} else {
		metric.Metrics.ChecksEnqueued.Inc()
	}
}
