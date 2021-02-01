package db_test

import (
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/db"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pipeline Factory", func() {
	var pipelineFactory db.PipelineFactory

	BeforeEach(func() {
		pipelineFactory = db.NewPipelineFactory(dbConn, lockFactory)
	})

	Describe("VisiblePipelines", func() {
		var (
			pipeline1 db.Pipeline
			pipeline2 db.Pipeline
			pipeline3 db.Pipeline
			pipeline4 db.Pipeline
		)

		BeforeEach(func() {
			err := defaultPipeline.Destroy()
			Expect(err).ToNot(HaveOccurred())

			team, err := teamFactory.CreateTeam(atc.Team{Name: "some-team"})
			Expect(err).ToNot(HaveOccurred())

			pipeline1, _, err = team.SavePipeline(atc.PipelineRef{Name: "fake-pipeline"}, atc.Config{
				Jobs: atc.JobConfigs{
					{Name: "job-name"},
				},
			}, db.ConfigVersion(1), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipeline1.Reload()).To(BeTrue())

			pipeline2, _, err = defaultTeam.SavePipeline(atc.PipelineRef{Name: "fake-pipeline-two"}, atc.Config{
				Jobs: atc.JobConfigs{
					{Name: "job-fake"},
				},
			}, db.ConfigVersion(1), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipeline2.Reload()).To(BeTrue())

			pipeline3, _, err = defaultTeam.SavePipeline(atc.PipelineRef{Name: "fake-pipeline-three"}, atc.Config{
				Jobs: atc.JobConfigs{
					{Name: "job-fake-two"},
				},
			}, db.ConfigVersion(1), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipeline3.Expose()).To(Succeed())
			Expect(pipeline3.Reload()).To(BeTrue())

			pipeline4, _, err = team.SavePipeline(atc.PipelineRef{Name: "fake-pipeline", InstanceVars: atc.InstanceVars{"branch": "master"}}, atc.Config{
				Jobs: atc.JobConfigs{
					{Name: "job-name"},
				},
			}, db.ConfigVersion(1), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipeline4.Reload()).To(BeTrue())
		})

		It("returns all pipelines visible for the given teams", func() {
			pipelines, err := pipelineFactory.VisiblePipelines([]string{"some-team"})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pipelines)).To(Equal(3))
			Expect(pipelines[0].Name()).To(Equal(pipeline1.Name()))
			Expect(pipelines[1].Name()).To(Equal(pipeline4.Name()))
			Expect(pipelines[2].Name()).To(Equal(pipeline3.Name()))
		})

		It("returns all pipelines visible when empty team name provided", func() {
			pipelines, err := pipelineFactory.VisiblePipelines([]string{""})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pipelines)).To(Equal(1))
			Expect(pipelines[0].Name()).To(Equal(pipeline3.Name()))
		})

		It("returns all pipelines visible when empty teams provided", func() {
			pipelines, err := pipelineFactory.VisiblePipelines([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pipelines)).To(Equal(1))
			Expect(pipelines[0].Name()).To(Equal(pipeline3.Name()))
		})

		It("returns all pipelines visible when nil teams provided", func() {
			pipelines, err := pipelineFactory.VisiblePipelines(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pipelines)).To(Equal(1))
			Expect(pipelines[0].Name()).To(Equal(pipeline3.Name()))
		})
	})

	Describe("AllPipelines", func() {
		var (
			pipeline1 db.Pipeline
			pipeline2 db.Pipeline
			pipeline3 db.Pipeline
			pipeline4 db.Pipeline
		)

		BeforeEach(func() {
			err := defaultPipeline.Destroy()
			Expect(err).ToNot(HaveOccurred())

			team, err := teamFactory.CreateTeam(atc.Team{Name: "some-team"})
			Expect(err).ToNot(HaveOccurred())

			pipeline2, _, err = team.SavePipeline(atc.PipelineRef{Name: "fake-pipeline-two"}, atc.Config{
				Jobs: atc.JobConfigs{
					{Name: "job-fake"},
				},
			}, db.ConfigVersion(1), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipeline2.Reload()).To(BeTrue())

			pipeline3, _, err = team.SavePipeline(atc.PipelineRef{Name: "fake-pipeline-three"}, atc.Config{
				Jobs: atc.JobConfigs{
					{Name: "job-fake-two"},
				},
			}, db.ConfigVersion(1), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipeline3.Expose()).To(Succeed())
			Expect(pipeline3.Reload()).To(BeTrue())

			pipeline1, _, err = defaultTeam.SavePipeline(atc.PipelineRef{Name: "fake-pipeline"}, atc.Config{
				Jobs: atc.JobConfigs{
					{Name: "job-name"},
				},
			}, db.ConfigVersion(1), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipeline1.Expose()).To(Succeed())
			Expect(pipeline1.Reload()).To(BeTrue())

			pipeline4, _, err = team.SavePipeline(atc.PipelineRef{Name: "fake-pipeline-two", InstanceVars: atc.InstanceVars{"branch": "master"}}, atc.Config{
				Jobs: atc.JobConfigs{
					{Name: "job-name"},
				},
			}, db.ConfigVersion(1), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipeline4.Reload()).To(BeTrue())

		})

		It("returns all pipelines ordered by team id -> ordering -> pipeline id", func() {
			pipelines, err := pipelineFactory.AllPipelines()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pipelines)).To(Equal(4))
			Expect(pipelines[0].Name()).To(Equal(pipeline1.Name()))
			Expect(pipelines[1].Name()).To(Equal(pipeline2.Name()))
			Expect(pipelines[2].Name()).To(Equal(pipeline4.Name()))
			Expect(pipelines[3].Name()).To(Equal(pipeline3.Name()))
		})
	})
})
