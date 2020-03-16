package e2e_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rudderlabs/rudder-server/config"
	"github.com/rudderlabs/rudder-server/jobsdb"
	"github.com/rudderlabs/rudder-server/tests/helpers"
	"github.com/rudderlabs/rudder-server/utils/misc"
	uuid "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var dbHandle *sql.DB
var gatewayDBPrefix string
var routerDBPrefix string
var batchRouterDBPrefix string
var dbPollFreqInS int = 1
var gatewayDBCheckBufferInS int = 15
var jobSuccessStatus string = "succeeded"

var _ = BeforeSuite(func() {
	var err error
	psqlInfo := jobsdb.GetConnectionString()
	dbHandle, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	gatewayDBPrefix = config.GetString("Gateway.CustomVal", "GW")
	routerDBPrefix = config.GetString("Router.CustomVal", "RT")
	batchRouterDBPrefix = config.GetString("Router.CustomVal", "BATCH_RT")
})

var _ = Describe("E2E", func() {

	Context("Without user sessions processing", func() {

		It("verify event stored in gateway 1. has right sourceId and writeKey, 2. enhanced with messageId, anonymousId, requestIP and receivedAt fields", func() {
			eventTypeMap := []string{"BATCH", "IDENTIFY", "GROUP", "TRACK", "SCREEN", "PAGE", "ALIAS"}
			for _, eventType := range eventTypeMap {
				time.Sleep(time.Second)
				initGatewayJobsCount := helpers.GetJobsCount(dbHandle, gatewayDBPrefix)

				//Source with WriteKey: 1Yc6YbOGg6U2E8rlj97ZdOawPyr has one S3 and one GA as destinations.
				//Source id for the above writekey is 1Yc6YceKLOcUYk8je9B0GQ65mmL
				switch eventType {
				case "BATCH":
					helpers.SendBatchRequest("1Yc6YbOGg6U2E8rlj97ZdOawPyr", helpers.RemoveKeyFromJSON(helpers.BatchPayload, "batch.0.messageId", "batch.0.anonymousId"))
				case "IDENTIFY":
					helpers.SendIdentifyRequest("1Yc6YbOGg6U2E8rlj97ZdOawPyr", helpers.RemoveKeyFromJSON(helpers.IdentifyPayload, "messageId", "anonymousId"))
				case "GROUP":
					helpers.SendGroupRequest("1Yc6YbOGg6U2E8rlj97ZdOawPyr", helpers.RemoveKeyFromJSON(helpers.GroupPayload, "messageId", "anonymousId"))
				case "TRACK":
					helpers.SendTrackRequest("1Yc6YbOGg6U2E8rlj97ZdOawPyr", helpers.RemoveKeyFromJSON(helpers.TrackPayload, "messageId", "anonymousId"))
				case "SCREEN":
					helpers.SendScreenRequest("1Yc6YbOGg6U2E8rlj97ZdOawPyr", helpers.RemoveKeyFromJSON(helpers.ScreenPayload, "messageId", "anonymousId"))
				case "PAGE":
					helpers.SendPageRequest("1Yc6YbOGg6U2E8rlj97ZdOawPyr", helpers.RemoveKeyFromJSON(helpers.PagePayload, "messageId", "anonymousId"))
				case "ALIAS":
					helpers.SendAliasRequest("1Yc6YbOGg6U2E8rlj97ZdOawPyr", helpers.RemoveKeyFromJSON(helpers.AliasPayload, "messageId", "anonymousId"))

				}

				Eventually(func() int {
					return helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
				}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(initGatewayJobsCount + 1))
				Eventually(func() map[string]bool {
					jobs := helpers.GetLatestJobs(dbHandle, gatewayDBPrefix, 1)
					var sourceId, messageId, anonymousId, writeKey, requestIP, receivedAt string
					for _, job := range jobs {
						sourceId = gjson.GetBytes(job.Parameters, "source_id").String()
						messageId = gjson.GetBytes(job.EventPayload, "batch.0.messageId").String()
						anonymousId = gjson.GetBytes(job.EventPayload, "batch.0.anonymousId").String()
						writeKey = gjson.GetBytes(job.EventPayload, "writeKey").String()
						requestIP = gjson.GetBytes(job.EventPayload, "requestIP").String()
						receivedAt = gjson.GetBytes(job.EventPayload, "receivedAt").String()
					}
					return map[string]bool{sourceId: true,
						"messageId":   len(messageId) > 0,
						"anonymousId": len(anonymousId) > 0,
						writeKey:      true,
						"requestIP":   len(requestIP) > 0,
						"receivedAt":  len(receivedAt) > 0}
				}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(map[string]bool{"1Yc6YceKLOcUYk8je9B0GQ65mmL": true,
					"messageId":                   true,
					"anonymousId":                 true,
					"1Yc6YbOGg6U2E8rlj97ZdOawPyr": true,
					"requestIP":                   true,
					"receivedAt":                  true}))
			}
		})

		It("verify health endpoint", func() {
			Eventually(func() bool {
				resp := helpers.SendHealthRequest()
				var err error
				res := gjson.GetBytes(resp, "goroutines")
				numOfGoroutines, err := strconv.Atoi(res.Str)
				if err != nil {
					panic(err)
				}

				if numOfGoroutines > 0 {
					resp, err = sjson.SetBytes(resp, "goroutines", true)
				} else {
					resp, err = sjson.SetBytes(resp, "goroutines", false)
				}
				if err != nil {
					panic(err)
				}

				// a map container to decode the JSON structure into
				c := make(map[string]interface{})

				// unmarschal JSON
				err = json.Unmarshal(resp, &c)
				if err != nil {
					panic(err)
				}

				return reflect.DeepEqual(map[string]interface{}{"server": "UP", "db": "UP", "acceptingEvents": "TRUE", "routingEvents": "TRUE", "mode": "NORMAL", "goroutines": true}, c)
			}, 2, dbPollFreqInS).Should(Equal(true))
		})

		It("verify version endpoint", func() {
			Eventually(func() bool {
				resp := helpers.SendVersionRequest()

				c := make(map[string]interface{})
				err := json.Unmarshal(resp, &c)
				if err != nil {
					panic(err)
				}

				keys := make([]string, len(c))

				i := 0
				// copy c's keys into k
				for s, _ := range c {
					keys[i] = s
					i++
				}
				return helpers.SameStringSlice([]string{"BuildDate", "BuiltBy", "Commit", "GitUrl", "Major", "Minor", "Patch", "Version"}, keys)
			}, 2, dbPollFreqInS).Should(Equal(true))
		})

		It("verify event is stored in both gateway and router db", func() {
			initGatewayJobsCount := helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			initialRouterJobsCount := helpers.GetJobsCount(dbHandle, routerDBPrefix)
			//initialRouterJobStatusCount := helpers.GetJobStatusCount(dbHandle, jobSuccessStatus, routerDBPrefix)

			//Source with WriteKey: 1Yc6YbOGg6U2E8rlj97ZdOawPyr has one S3 and one GA as destinations.
			helpers.SendEventRequest(helpers.EventOptsT{
				WriteKey: "1Yc6YbOGg6U2E8rlj97ZdOawPyr",
			})

			// wait for some seconds for events to be processed by gateway
			time.Sleep(6 * time.Second)
			Eventually(func() int {
				return helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(initGatewayJobsCount + 1))
			Eventually(func() int {
				return helpers.GetJobsCount(dbHandle, routerDBPrefix)
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialRouterJobsCount + 1))
		})

		//Source with WriteKey: 1YcF00dWZXGjWpSIkfFnbGuI6OI has one GA and one AMPLITUDE as destinations.
		It("should create router job for both GA and AM for single event request", func() {
			initialRouterJobsCount := helpers.GetJobsCount(dbHandle, routerDBPrefix)
			//initialRouterJobStatusCount := helpers.GetJobStatusCount(dbHandle, jobSuccessStatus, routerDBPrefix)
			helpers.SendEventRequest(helpers.EventOptsT{
				WriteKey: "1YcF00dWZXGjWpSIkfFnbGuI6OI",
			})
			// wait for some seconds for events to be processed by gateway
			time.Sleep(6 * time.Second)
			Eventually(func() int {
				return helpers.GetJobsCount(dbHandle, routerDBPrefix)
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialRouterJobsCount + 2))
			Eventually(func() []string {
				jobs := helpers.GetLatestJobs(dbHandle, routerDBPrefix, 2)
				customVals := []string{}
				for _, job := range jobs {
					customVals = append(customVals, job.CustomVal)
				}
				return customVals
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(SatisfyAny(Or(BeEquivalentTo([]string{"GA", "AM"}), BeEquivalentTo([]string{"AM", "GA"}))))
		})

		It("should not create job with invalid write key", func() {
			initGatewayJobsCount := helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			helpers.SendEventRequest(helpers.EventOptsT{
				WriteKey: "invalid_write_key",
			})
			Eventually(func() int {
				return helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			}, 2, dbPollFreqInS).Should(Equal(initGatewayJobsCount))
		})

		It("should maintain order of events", func() {
			initialRouterJobsCount := helpers.GetJobsCount(dbHandle, routerDBPrefix)
			numOfTestEvents := 100
			for i := 1; i <= numOfTestEvents; i++ {
				helpers.SendEventRequest(helpers.EventOptsT{
					GaVal: i,
				})
			}
			// wait for some seconds for events to be processed by gateway
			time.Sleep(6 * time.Second)
			Eventually(func() int {
				return helpers.GetJobsCount(dbHandle, routerDBPrefix)
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialRouterJobsCount + numOfTestEvents))
			jobs := helpers.GetLatestJobs(dbHandle, routerDBPrefix, numOfTestEvents)
			for index, _ := range jobs {
				if index == 0 {
					continue
				}
				result1, _ := strconv.Atoi(gjson.Get(string(jobs[index].EventPayload), "params.ev").String())
				result2, _ := strconv.Atoi(gjson.Get(string(jobs[index-1].EventPayload), "params.ev").String())
				Expect(result1).Should(BeNumerically("<", result2))
			}
		})

		It("should dedup duplicate events", func() {
			sampleID := uuid.NewV4().String()

			initGatewayJobsCount := helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			helpers.SendEventRequest(helpers.EventOptsT{
				MessageID: sampleID,
			})
			Eventually(func() int {
				return helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(initGatewayJobsCount + 1))

			// send 2 events and verify event with prev messageID is dropped
			currentGatewayJobsCount := helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			helpers.SendEventRequest(helpers.EventOptsT{
				MessageID: sampleID,
			})
			helpers.SendEventRequest(helpers.EventOptsT{
				MessageID: uuid.NewV4().String(),
			})
			Consistently(func() int {
				return helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(currentGatewayJobsCount + 1))
		})

		It("should dedup duplicate events only till specified TTL", func() {
			sampleID := uuid.NewV4().String()

			initGatewayJobsCount := helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			helpers.SendEventRequest(helpers.EventOptsT{
				MessageID: sampleID,
			})
			helpers.SendEventRequest(helpers.EventOptsT{
				MessageID: sampleID,
			})
			Eventually(func() int {
				return helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(initGatewayJobsCount + 1))

			// send 2 events and verify event with prev messageID is dropped
			currentGatewayJobsCount := helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			// wait for 5 seconds for messageID to exceed its TTL
			time.Sleep(5 * time.Second)
			helpers.SendEventRequest(helpers.EventOptsT{
				MessageID: sampleID,
			})
			Eventually(func() int {
				return helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(currentGatewayJobsCount + 1))

		})
		It("should enhance events with time related fields and metadata", func() {
			initGatewayJobsCount := helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			// initialRouterJobsCount := helpers.GetJobsCount(dbHandle, routerDBPrefix)
			eventWriteKey := "1Yc6YbOGg6U2E8rlj97ZdOawPyr"
			//Source with WriteKey: 1Yc6YbOGg6U2E8rlj97ZdOawPyr has one S3 and one GA as destinations.
			helpers.SendEventRequest(helpers.EventOptsT{
				WriteKey: eventWriteKey,
			})

			// wait for some seconds for events to be processed by gateway
			time.Sleep(6 * time.Second)
			Eventually(func() int {
				return helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(initGatewayJobsCount + 1))
			jobs := helpers.GetJobs(dbHandle, gatewayDBPrefix, 1)
			if len(jobs) != 1 {
				panic("Jobs count mismatch")
			}
			job := jobs[0]
			requestIP := gjson.Get(string(job.EventPayload), "requestIP").String()
			writeKey := gjson.Get(string(job.EventPayload), "writeKey").String()
			receivedAt := gjson.Get(string(job.EventPayload), "receivedAt").String()

			Expect(writeKey).To(Equal(eventWriteKey))
			Expect(requestIP).NotTo(BeNil()) //TODO: Figure out a way to test requestIP
			Expect(time.Parse(misc.RFC3339Milli, receivedAt)).Should(BeTemporally("~", time.Now(), 10*time.Second))

		})
		It("should send correct response headers to support CORS", func() {
			originURL := "http://example.com/"
			serverIP := "http://localhost:8080/v1/track"
			req, _ := http.NewRequest("POST", serverIP, bytes.NewBuffer([]byte("{}")))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Origin", originURL)
			req.SetBasicAuth("1Yc6YbOGg6U2E8rlj97ZdOawPyr", "")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			Expect(resp.Header.Get("Access-Control-Allow-Credentials")).To(Equal("true"))
			Expect(resp.Header.Get("Access-Control-Allow-Origin")).To(Equal(originURL))
		})

		It("should handle different cases in user transformation functions", func() {
			gatewayDBCheckBufferInS = 10
			sourceID := "1YyeVLcPH3rKK3ehTjTUfQyxiRS"
			initGatewayJobsCount := helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			initialRouterJobsCount := map[string]int{
				"1YyedMM9RCUJrUUywWXkJWKzQiW": helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, "1YyedMM9RCUJrUUywWXkJWKzQiW"), // Empty array
				"1YysQJV2ZXgp8wgXRhU3mmwHNti": helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, "1YysQJV2ZXgp8wgXRhU3mmwHNti"), // Access element outside array
				"1YytOBIeLOdPCxrXWF8DytawB9Z": helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, "1YytOBIeLOdPCxrXWF8DytawB9Z"), // Events in wrong format
				"1Yytf8vs6vxp5TsLW58D44QPKeT": helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, "1Yytf8vs6vxp5TsLW58D44QPKeT"), // Memory Eater
				"1YyuocAqn8Q78fqwvBSDks7BVxO": helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, "1YyuocAqn8Q78fqwvBSDks7BVxO"), // Api call
			}
			initialBatchRouterJobsCount := map[string]int{
				"1YyrxAh3Q7FMktGV5GWPDAsldKw": helpers.GetJobsCountForSourceAndDestination(dbHandle, batchRouterDBPrefix, sourceID, "1YyrxAh3Q7FMktGV5GWPDAsldKw"), // Infinite loop
				"1YysY9LMizekwLHwM06lmT7Oo47": helpers.GetJobsCountForSourceAndDestination(dbHandle, batchRouterDBPrefix, sourceID, "1YysY9LMizekwLHwM06lmT7Oo47"), // Syntax error
				"1YytSagaEVPTK1DLXJ46A9RQBXA": helpers.GetJobsCountForSourceAndDestination(dbHandle, batchRouterDBPrefix, sourceID, "1YytSagaEVPTK1DLXJ46A9RQBXA"), // Single event outside array
				"1Yytj13urVx2WmDjOcWBHdZCRGA": helpers.GetJobsCountForSourceAndDestination(dbHandle, batchRouterDBPrefix, sourceID, "1Yytj13urVx2WmDjOcWBHdZCRGA"), // Api call
			}
			//Source with WriteKey: 1YyeVOLDReXPNYORjDlvE7PM1Jw has multiple destinations with diffent user-transformation for differnet destiantions.
			helpers.SendEventRequest(helpers.EventOptsT{
				WriteKey: "1YyeVOLDReXPNYORjDlvE7PM1Jw",
			})

			// wait for some seconds for events to be processed by gateway
			time.Sleep(6 * time.Second)
			Eventually(func() int {
				return helpers.GetJobsCount(dbHandle, gatewayDBPrefix)
			}, gatewayDBCheckBufferInS, dbPollFreqInS).Should(Equal(initGatewayJobsCount + 1))

			routerDBCheckBufferInS := 10

			destinationID := "1YyedMM9RCUJrUUywWXkJWKzQiW"
			Consistently(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, destinationID)
			}, routerDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialRouterJobsCount[destinationID]))

			destinationID = "1YysQJV2ZXgp8wgXRhU3mmwHNti"

			Consistently(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, destinationID)
			}, routerDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialRouterJobsCount[destinationID]))

			destinationID = "1YytOBIeLOdPCxrXWF8DytawB9Z"
			Consistently(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, destinationID)
			}, routerDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialRouterJobsCount[destinationID]))

			destinationID = "1Yytf8vs6vxp5TsLW58D44QPKeT"
			Consistently(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, destinationID)
			}, routerDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialRouterJobsCount[destinationID]))

			destinationID = "1YyuocAqn8Q78fqwvBSDks7BVxO"
			Eventually(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, destinationID)
			}, routerDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialRouterJobsCount[destinationID] + 1))
			Consistently(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, routerDBPrefix, sourceID, destinationID)
			}, routerDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialRouterJobsCount[destinationID] + 1))

			batchRouterDBCheckBufferInS := 10

			// For the destination with infinite loop
			destinationID = "1YyrxAh3Q7FMktGV5GWPDAsldKw"
			Consistently(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, batchRouterDBPrefix, sourceID, destinationID)
			}, batchRouterDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialBatchRouterJobsCount[destinationID]), "User transformation has an Infinite Loop")

			destinationID = "1YysY9LMizekwLHwM06lmT7Oo47"
			Consistently(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, batchRouterDBPrefix, sourceID, destinationID)
			}, batchRouterDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialBatchRouterJobsCount[destinationID]), "User transformation has syntax error")

			destinationID = "1YytSagaEVPTK1DLXJ46A9RQBXA"
			Consistently(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, batchRouterDBPrefix, sourceID, destinationID)
			}, batchRouterDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialBatchRouterJobsCount[destinationID]), "User transformation returns a single event outside array")

			// For the destination with API call UT
			destinationID = "1Yytj13urVx2WmDjOcWBHdZCRGA"
			Eventually(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, batchRouterDBPrefix, "1YyeVLcPH3rKK3ehTjTUfQyxiRS", destinationID)
			}, batchRouterDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialBatchRouterJobsCount[destinationID]+1), "User transformation with API call")
			Consistently(func() int {
				return helpers.GetJobsCountForSourceAndDestination(dbHandle, batchRouterDBPrefix, "1YyeVLcPH3rKK3ehTjTUfQyxiRS", destinationID)
			}, batchRouterDBCheckBufferInS, dbPollFreqInS).Should(Equal(initialBatchRouterJobsCount[destinationID]+1), "User transformation with API call")
		})
	})

})
