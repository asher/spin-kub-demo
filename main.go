package main

import (
  "encoding/json"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "math/rand"
  "net/http"
  "os"
  "regexp"
  "strings"
  "time"

  "golang.org/x/net/context"
  "golang.org/x/oauth2/google"

  "google.golang.org/api/monitoring/v3"

  "cloud.google.com/go/compute/metadata"
)

// TODO cli or env configurable
const errorMetricType = "custom.googleapis.com/workshop/canary/request/errors"
const randomMetricType1 = "custom.googleapis.com/workshop/canary/request/random1"
const randomMetricType2 = "custom.googleapis.com/workshop/canary/request/random2"
const projectID = "qcon-2017-workshop"

func index(w http.ResponseWriter, r *http.Request) {
  fmt.Printf("Handling %+v\n", r)
  bs, err := ioutil.ReadFile("/app/content/index.html")

  if err != nil {
    fmt.Printf("Couldn't read index.html: %v", err)
    os.Exit(1)
  }

  io.WriteString(w, string(bs[:]))
}

func projectResource(projectID string) string {
  return "projects/" + projectID
}

func createService(ctx context.Context) (*monitoring.Service, error) {
  hc, err := google.DefaultClient(ctx, monitoring.MonitoringScope)
  if err != nil {
    return nil, err
  }
  s, err := monitoring.New(hc)
  if err != nil {
    return nil, err
  }
  return s, nil
}

// writeTimeSeriesValue writes a value for the custom metric created
func writeTimeSeriesValue(s *monitoring.Service, projectID, metricType string, cluster string, sg string) error {
  instanceId, _ := metadata.InstanceID() // subject to change after start due to live migration
  zone, _ := metadata.Zone()

  now := time.Now().UTC().Format(time.RFC3339Nano)
  randVal := rand.Float64()

  if (strings.Contains(metricType, "error")) {
    match, _ := regexp.MatchString("baseline", cluster)
    if match {
      randVal *= 0.1
    } else {
      randVal *= 2
    }
  } else {
    randVal *= 0.5
  }

  timeseries := monitoring.TimeSeries{
    Metric: &monitoring.Metric{
      Type: metricType,
      Labels: map[string]string{
        "cluster":     cluster,
        "servergroup": sg,
      },
    },
    Resource: &monitoring.MonitoredResource{
      Labels: map[string]string{
        "instance_id": instanceId,
        "zone":        zone,
      },
      Type: "gce_instance",
    },
    Points: []*monitoring.Point{
      {
        Interval: &monitoring.TimeInterval{
          StartTime: now,
          EndTime:   now,
        },
        Value: &monitoring.TypedValue{
          DoubleValue: &randVal,
        },
      },
    },
  }

  createTimeseriesRequest := monitoring.CreateTimeSeriesRequest{
    TimeSeries: []*monitoring.TimeSeries{&timeseries},
  }

  log.Printf("writeTimeseriesRequest: %s\n", formatResource(createTimeseriesRequest))
  _, err := s.Projects.TimeSeries.Create(projectResource(projectID), &createTimeseriesRequest).Do()
  if err != nil {
    log.Printf("Could not write time series value, %v \n", err)
  }
  return nil
}

// formatResource marshals a response object as JSON.
func formatResource(resource interface{}) []byte {
  b, err := json.MarshalIndent(resource, "", "    ")
  if err != nil {
    panic(err)
  }
  return b
}

func metrics() {
  hostname := os.Getenv("HOSTNAME")
  clusteRe, _ := regexp.Compile(`^([\w\-]+)\-v\d+\-\w+$`)
  sgRe, _ := regexp.Compile(`^([\w\-]+\-v\d+)\-\w+$`)
  cluster := clusteRe.FindStringSubmatch(hostname)[1]
  sg := sgRe.FindStringSubmatch(hostname)[1]

  ctx := context.Background()
  s, err := createService(ctx)
  if err != nil {
    log.Fatal(err)
  }

  for {
    writeTimeSeriesValue(s, projectID, errorMetricType, cluster, sg)
    writeTimeSeriesValue(s, projectID, randomMetricType1, cluster, sg)
    writeTimeSeriesValue(s, projectID, randomMetricType2, cluster, sg)
    time.Sleep(time.Second * 10)
  }
}

func main() {
  for _, e := range os.Environ() {
    fmt.Println(e)
  }
  go metrics()
  http.HandleFunc("/", index)
  port := ":80"
  fmt.Printf("Starting to service on port %s\n", port)
  http.ListenAndServe(port, nil)
}
