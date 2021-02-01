package concourse_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/go-concourse/concourse"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("ATC Handler Jobs", func() {
	Describe("team.ListJobs", func() {
		var (
			expectedJobs  []atc.Job
			expectedURL   = "/api/v1/teams/some-team/pipelines/mypipeline/jobs"
			expectedQuery = "vars.branch=%22master%22"
			pipelineRef   = atc.PipelineRef{Name: "mypipeline", InstanceVars: atc.InstanceVars{"branch": "master"}}
		)

		BeforeEach(func() {
			expectedJobs = []atc.Job{
				{
					Name:      "myjob-1",
					NextBuild: nil,
				},
				{
					Name:      "myjob-2",
					NextBuild: nil,
				},
			}

			atcServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", expectedURL, expectedQuery),
					ghttp.RespondWithJSONEncoded(http.StatusOK, expectedJobs),
				),
			)
		})

		It("returns jobs that belong to the pipeline", func() {
			pipelines, err := team.ListJobs(pipelineRef)
			Expect(err).NotTo(HaveOccurred())
			Expect(pipelines).To(Equal(expectedJobs))
		})
	})

	Describe("client.ListAllJobs", func() {
		var expectedJobs []atc.Job

		BeforeEach(func() {
			expectedURL := "/api/v1/jobs"

			expectedJobs = []atc.Job{
				{
					Name:      "myjob-1",
					NextBuild: nil,
				},
				{
					Name:      "myjob-2",
					NextBuild: nil,
				},
			}

			atcServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", expectedURL),
					ghttp.RespondWithJSONEncoded(http.StatusOK, expectedJobs),
				),
			)
		})

		It("returns all jobs that belong to the account", func() {
			jobs, err := client.ListAllJobs()
			Expect(err).NotTo(HaveOccurred())
			Expect(jobs).To(Equal(expectedJobs))
		})
	})

	Describe("Job", func() {
		var (
			expectedJob atc.Job
			expectedURL = "/api/v1/teams/some-team/pipelines/mypipeline/jobs/myjob"
			queryParams = "vars.branch=%22master%22"
			pipelineRef = atc.PipelineRef{Name: "mypipeline", InstanceVars: atc.InstanceVars{"branch": "master"}}
		)

		Context("when job exists", func() {
			BeforeEach(func() {
				expectedJob = atc.Job{
					Name:      "myjob",
					NextBuild: nil,
					FinishedBuild: &atc.Build{
						ID:      123,
						Name:    "mybuild",
						Status:  "succeeded",
						JobName: "myjob",
						APIURL:  "api/v1/teams/some-team/builds/123",
					},
					Inputs: []atc.JobInput{
						{
							Name:     "myfirstinput",
							Resource: "myfirstinput",
							Passed:   []string{"rc"},
							Trigger:  true,
						},
						{
							Name:     "mysecondinput",
							Resource: "mysecondinput",
							Passed:   []string{"rc"},
							Trigger:  true,
						},
					},
					Outputs: []atc.JobOutput{
						{
							Name:     "myfirstoutput",
							Resource: "myfirstoutput",
						},
						{
							Name:     "mysecoundoutput",
							Resource: "mysecoundoutput",
						},
					},
					Groups: []string{"mygroup"},
				}

				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL, queryParams),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedJob),
					),
				)
			})

			It("returns the given job for that pipeline", func() {
				job, found, err := team.Job(pipelineRef, "myjob")
				Expect(err).NotTo(HaveOccurred())
				Expect(job).To(Equal(expectedJob))
				Expect(found).To(BeTrue())
			})
		})

		Context("when job does not exist", func() {
			BeforeEach(func() {
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL, queryParams),
						ghttp.RespondWith(http.StatusNotFound, ""),
					),
				)
			})

			It("returns false and no error", func() {
				_, found, err := team.Job(pipelineRef, "myjob")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("JobBuilds", func() {
		var (
			expectedBuilds []atc.Build
			expectedURL    = "/api/v1/teams/some-team/pipelines/mypipeline/jobs/myjob/builds"
			expectedQuery  []string
			pipelineRef    = atc.PipelineRef{Name: "mypipeline", InstanceVars: atc.InstanceVars{"branch": "master"}}
		)

		BeforeEach(func() {
			expectedQuery = []string{"vars.branch=%22master%22"}
		})

		JustBeforeEach(func() {
			expectedBuilds = []atc.Build{
				{
					Name: "some-build",
				},
				{
					Name: "some-other-build",
				},
			}

			atcServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", expectedURL, strings.Join(expectedQuery, "&")),
					ghttp.RespondWithJSONEncoded(http.StatusOK, expectedBuilds),
				),
			)
		})

		Context("when from, to, and limit are 0", func() {
			It("calls to get all builds", func() {
				builds, _, found, err := team.JobBuilds(pipelineRef, "myjob", concourse.Page{})
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(builds).To(Equal(expectedBuilds))
			})
		})

		Context("when from is specified", func() {
			BeforeEach(func() {
				expectedQuery = append(expectedQuery, "from=24")
			})

			It("calls to get all builds from that id", func() {
				builds, _, found, err := team.JobBuilds(pipelineRef, "myjob", concourse.Page{From: 24})
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(builds).To(Equal(expectedBuilds))
			})

			Context("and limit is specified", func() {
				BeforeEach(func() {
					expectedQuery = append(expectedQuery, "limit=5")
				})

				It("appends limit to the url", func() {
					builds, _, found, err := team.JobBuilds(pipelineRef, "myjob", concourse.Page{From: 24, Limit: 5})
					Expect(err).NotTo(HaveOccurred())
					Expect(found).To(BeTrue())
					Expect(builds).To(Equal(expectedBuilds))
				})
			})
		})

		Context("when to is specified", func() {
			BeforeEach(func() {
				expectedQuery = append(expectedQuery, "to=26")
			})

			It("calls to get all builds to that id", func() {
				builds, _, found, err := team.JobBuilds(pipelineRef, "myjob", concourse.Page{To: 26})
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(builds).To(Equal(expectedBuilds))
			})

			Context("and limit is specified", func() {
				BeforeEach(func() {
					expectedQuery = append(expectedQuery, "limit=15")
				})

				It("appends limit to the url", func() {
					builds, _, found, err := team.JobBuilds(pipelineRef, "myjob", concourse.Page{To: 26, Limit: 15})
					Expect(err).NotTo(HaveOccurred())
					Expect(found).To(BeTrue())
					Expect(builds).To(Equal(expectedBuilds))
				})
			})
		})

		Context("when from and to are both specified", func() {
			BeforeEach(func() {
				expectedQuery = append(expectedQuery, "to=26", "from=24")
			})

			It("sends both the from and the to", func() {
				builds, _, found, err := team.JobBuilds(pipelineRef, "myjob", concourse.Page{From: 24, To: 26})
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(builds).To(Equal(expectedBuilds))
			})
		})

		Context("when the server returns an error", func() {
			BeforeEach(func() {
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL, strings.Join(expectedQuery, "&")),
						ghttp.RespondWith(http.StatusInternalServerError, ""),
					),
				)
			})

			It("returns false and an error", func() {
				_, _, found, err := team.JobBuilds(pipelineRef, "myjob", concourse.Page{})
				Expect(err).To(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when the server returns not found", func() {
			BeforeEach(func() {
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL, strings.Join(expectedQuery, "&")),
						ghttp.RespondWith(http.StatusNotFound, ""),
					),
				)
			})

			It("returns false and no error", func() {
				_, _, found, err := team.JobBuilds(pipelineRef, "myjob", concourse.Page{})
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("pagination data", func() {
			Context("with a link header", func() {
				BeforeEach(func() {
					atcServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", expectedURL, strings.Join(expectedQuery, "&")),
							ghttp.RespondWithJSONEncoded(http.StatusOK, expectedBuilds, http.Header{
								"Link": []string{
									`<http://some-url.com/api/v1/teams/some-team/pipelines/some-pipeline/jobs/some-job/builds?from=452&limit=123>; rel="previous"`,
									`<http://some-url.com/api/v1/teams/some-team/pipelines/some-pipeline/jobs/some-job/builds?to=254&limit=456>; rel="next"`,
								},
							}),
						),
					)
				})

				It("returns the pagination data from the header", func() {
					_, pagination, _, err := team.JobBuilds(pipelineRef, "myjob", concourse.Page{})
					Expect(err).ToNot(HaveOccurred())

					Expect(pagination.Previous).To(Equal(&concourse.Page{From: 452, Limit: 123}))
					Expect(pagination.Next).To(Equal(&concourse.Page{To: 254, Limit: 456}))
				})
			})
		})

		Context("without a link header", func() {
			BeforeEach(func() {
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL, strings.Join(expectedQuery, "&")),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedBuilds, http.Header{}),
					),
				)
			})

			It("returns pagination data with nil pages", func() {
				_, pagination, _, err := team.JobBuilds(pipelineRef, "myjob", concourse.Page{})
				Expect(err).ToNot(HaveOccurred())

				Expect(pagination.Previous).To(BeNil())
				Expect(pagination.Next).To(BeNil())
			})
		})
	})

	Describe("PauseJob", func() {
		var (
			expectedStatus int
			pipelineName   = "banana"
			jobName        = "disjob"
			expectedURL    = fmt.Sprintf("/api/v1/teams/some-team/pipelines/%s/jobs/%s/pause", pipelineName, jobName)
			expectedQuery  = "vars.branch=%22master%22"
			pipelineRef    = atc.PipelineRef{Name: pipelineName, InstanceVars: atc.InstanceVars{"branch": "master"}}
		)

		JustBeforeEach(func() {
			atcServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", expectedURL, expectedQuery),
					ghttp.RespondWith(expectedStatus, nil),
				),
			)
		})

		Context("when the job exists and there are no issues", func() {
			BeforeEach(func() {
				expectedStatus = http.StatusOK
			})

			It("calls the pause job and returns no error", func() {
				Expect(func() {
					paused, err := team.PauseJob(pipelineRef, jobName)
					Expect(err).NotTo(HaveOccurred())
					Expect(paused).To(BeTrue())
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(1))
			})
		})

		Context("when the pause job call fails", func() {
			BeforeEach(func() {
				expectedStatus = http.StatusInternalServerError
			})

			It("calls the pause job and returns an error", func() {
				Expect(func() {
					paused, err := team.PauseJob(pipelineRef, jobName)
					Expect(err).To(HaveOccurred())
					Expect(paused).To(BeFalse())
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(1))
			})
		})

		Context("when the job does not exist", func() {
			BeforeEach(func() {
				expectedStatus = http.StatusNotFound
			})

			It("calls the pause job and returns an error", func() {
				Expect(func() {
					paused, err := team.PauseJob(pipelineRef, jobName)
					Expect(err).ToNot(HaveOccurred())
					Expect(paused).To(BeFalse())
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(1))
			})
		})
	})

	Describe("UnpauseJob", func() {
		var (
			expectedStatus int
			pipelineName   = "banana"
			jobName        = "disjob"
			expectedURL    = fmt.Sprintf("/api/v1/teams/some-team/pipelines/%s/jobs/%s/unpause", pipelineName, jobName)
			expectedQuery  = "vars.branch=%22master%22"
			pipelineRef    = atc.PipelineRef{Name: pipelineName, InstanceVars: atc.InstanceVars{"branch": "master"}}
		)

		JustBeforeEach(func() {
			atcServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", expectedURL, expectedQuery),
					ghttp.RespondWith(expectedStatus, nil),
				),
			)
		})

		Context("when the job exists and there are no issues", func() {
			BeforeEach(func() {
				expectedStatus = http.StatusOK
			})

			It("calls the pause job and returns no error", func() {
				Expect(func() {
					paused, err := team.UnpauseJob(pipelineRef, jobName)
					Expect(err).NotTo(HaveOccurred())
					Expect(paused).To(BeTrue())
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(1))
			})
		})

		Context("when the pause job call fails", func() {
			BeforeEach(func() {
				expectedStatus = http.StatusInternalServerError
			})

			It("calls the pause job and returns an error", func() {
				Expect(func() {
					paused, err := team.UnpauseJob(pipelineRef, jobName)
					Expect(err).To(HaveOccurred())
					Expect(paused).To(BeFalse())
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(1))
			})
		})

		Context("when the job does not exist", func() {
			BeforeEach(func() {
				expectedStatus = http.StatusNotFound
			})

			It("calls the pause job and returns an error", func() {
				Expect(func() {
					paused, err := team.UnpauseJob(pipelineRef, jobName)
					Expect(err).ToNot(HaveOccurred())
					Expect(paused).To(BeFalse())
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(1))
			})
		})
	})

	Describe("ScheduleJob", func() {
		var (
			expectedStatus int
			pipelineName   = "banana"
			jobName        = "disjob"
			expectedURL    = fmt.Sprintf("/api/v1/teams/some-team/pipelines/%s/jobs/%s/schedule", pipelineName, jobName)
			expectedQuery  = "vars.branch=%22master%22"
			pipelineRef    = atc.PipelineRef{Name: pipelineName, InstanceVars: atc.InstanceVars{"branch": "master"}}
		)

		JustBeforeEach(func() {
			atcServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", expectedURL, expectedQuery),
					ghttp.RespondWith(expectedStatus, nil),
				),
			)
		})

		Context("when the job exists and there are no issues", func() {
			BeforeEach(func() {
				expectedStatus = http.StatusOK
			})

			It("calls the schedule job and returns no error", func() {
				Expect(func() {
					requested, err := team.ScheduleJob(pipelineRef, jobName)
					Expect(err).NotTo(HaveOccurred())
					Expect(requested).To(BeTrue())
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(1))
			})
		})

		Context("when the schedule job call fails", func() {
			BeforeEach(func() {
				expectedStatus = http.StatusInternalServerError
			})

			It("calls the schedule job and returns an error", func() {
				Expect(func() {
					requested, err := team.ScheduleJob(pipelineRef, jobName)
					Expect(err).To(HaveOccurred())
					Expect(requested).To(BeFalse())
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(1))
			})
		})

		Context("when the job does not exist", func() {
			BeforeEach(func() {
				expectedStatus = http.StatusNotFound
			})

			It("calls the schedule job and returns an error", func() {
				Expect(func() {
					requested, err := team.ScheduleJob(pipelineRef, jobName)
					Expect(err).ToNot(HaveOccurred())
					Expect(requested).To(BeFalse())
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(1))
			})
		})
	})

	Describe("Clear Job Task Cache", func() {
		var (
			expectedURL   string
			requestMethod string
			expectedQuery []string
			pipelineRef   = atc.PipelineRef{Name: "mypipeline", InstanceVars: atc.InstanceVars{"branch": "master"}}
		)

		BeforeEach(func() {
			requestMethod = "DELETE"
			expectedQuery = []string{"vars.branch=%22master%22"}
		})

		Context("when job step exists", func() {
			JustBeforeEach(func() {
				expectedURL = fmt.Sprint("/api/v1/teams/some-team/pipelines/mypipeline/jobs/myjob/tasks/mystep/cache")
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(requestMethod, expectedURL, strings.Join(expectedQuery, "&")),
						ghttp.RespondWithJSONEncoded(http.StatusOK, atc.ClearTaskCacheResponse{CachesRemoved: 1}),
					),
				)
			})

			Context("when no cache path is given", func() {
				It("succeeds", func() {
					Expect(func() {
						numDeleted, err := team.ClearTaskCache(pipelineRef, "myjob", "mystep", "")
						Expect(err).NotTo(HaveOccurred())
						Expect(numDeleted).To(Equal(int64(1)))
					}).To(Change(func() int {
						return len(atcServer.ReceivedRequests())
					}).By(1))
				})
			})

			Context("when a cache path is given", func() {
				BeforeEach(func() {
					expectedQuery = append(expectedQuery, "cache_path=mycachepath")
				})

				Context("when the cache path exists", func() {
					It("succeeds", func() {
						Expect(func() {
							numDeleted, err := team.ClearTaskCache(pipelineRef, "myjob", "mystep", "mycachepath")
							Expect(err).NotTo(HaveOccurred())
							Expect(numDeleted).To(Equal(int64(1)))
						}).To(Change(func() int {
							return len(atcServer.ReceivedRequests())
						}).By(1))
					})
				})
			})

		})

		Context("when job step does not exist", func() {
			BeforeEach(func() {
				expectedURL = fmt.Sprint("/api/v1/teams/some-team/pipelines/mypipeline/jobs/myjob/tasks/my-nonexistent-step/cache")
				expectedQuery = append(expectedQuery, "cache_path=mycachepath")
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(requestMethod, expectedURL, strings.Join(expectedQuery, "&")),
						ghttp.RespondWithJSONEncoded(http.StatusOK, atc.ClearTaskCacheResponse{CachesRemoved: 0}),
					),
				)
			})

			It("returns that 0 caches were deleted", func() {
				Expect(func() {
					numDeleted, err := team.ClearTaskCache(pipelineRef, "myjob", "my-nonexistent-step", "mycachepath")
					Expect(err).NotTo(HaveOccurred())
					Expect(numDeleted).To(BeZero())
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(1))
			})
		})
	})

})
