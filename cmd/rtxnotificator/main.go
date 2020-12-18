// Loads env variables from env. file and runs the program.
package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"math/rand"
	"os"
	"rtx-notificator/apiservice"
	"rtx-notificator/apiservice/nvidiade"
	"rtx-notificator/msgservice"
	"rtx-notificator/msgservice/twilio"
	"strconv"
	"strings"
	"time"
)

const (
	ENV_POLL_INTERVAL = "POLL_INTERVAL"
)

const (
	LFakeT = "FakeTest"
	LLiveT = "LiveTest"
	LError = "Error"
	LInfo  = "Info"
	LPoll  = "Poll"
)

func main() {
	// Load environment variables from .env file.
	godotenv.Load()

	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	var sid, token, sidTest, tokenTest, from string
	var to []string
	var pollInterval time.Duration

	fakeTestFlag := flag.Bool("faketest", false, "Test Twilio SMS api with test credentials.")
	liveTestFlag := flag.Bool("livetest", false, "Test Twilio SMS api with real credentials. The client will send SMS to recipients - you will be charged for these.")
	flagPoll := flag.Int("pollint", 1000, "Interval at which the availability of the GPUs are checked in [ms]")
	envFlag := flag.Bool("env", false, fmt.Sprintf("These environment variables are read: \n\t%v -> Twilio sender tel. number\n\t%v -> Space separated recipient tel. numbers\n\t%v -> Twilio SID\n\t%v -> Twilio authentication token\n\t%v -> Twilio test SID\n  \t%v -> Twilio test authentication token\n \t%v -> Polling interval in [ms]", twilio.ENV_TWILIO_NUM_FROM, twilio.ENV_TWILIO_NUM_TO, twilio.ENV_TWILIO_SID, twilio.ENV_TWILIO_AUTH_TOKEN, twilio.ENV_TWILIO_TEST_SID, twilio.ENV_TWILIO_TEST_AUTH_TOKEN, ENV_POLL_INTERVAL))
	flag.Parse()

	pollInterval = time.Duration(*flagPoll) * time.Millisecond

	if *envFlag {
		// For live test.
		sid = os.Getenv(twilio.ENV_TWILIO_SID)
		token = os.Getenv(twilio.ENV_TWILIO_AUTH_TOKEN)
		sidTest = os.Getenv(twilio.ENV_TWILIO_TEST_SID)
		tokenTest = os.Getenv(twilio.ENV_TWILIO_TEST_AUTH_TOKEN)

		// For fake test.
		from = os.Getenv(twilio.ENV_TWILIO_NUM_FROM)
		to = strings.Split(os.Getenv(twilio.ENV_TWILIO_NUM_TO), " ")

		val, err := strconv.Atoi(os.Getenv(ENV_POLL_INTERVAL))
		pollInterval = time.Duration(val) * time.Millisecond
		if err != nil {
			printf(LError, "Invalid %v environment variable, define it as a number", ENV_POLL_INTERVAL)
			os.Exit(1)
		}
	}

	var hasFakeCred, hasLiveCred, hasRecipients, hasSender bool

	if sid != "" || token != "" {
		hasLiveCred = true
	}

	if sidTest != "" || tokenTest != "" {
		hasFakeCred = true
	}

	if len(to) > 0 {
		hasRecipients = true
	}

	if from != "" {
		hasSender = true
	}

	if !hasSender || !hasRecipients {
		printf(LError, "Please provide the sender and recipients tel. numbers.")
		os.Exit(1)
	}

	// Testing with fake credentials.
	if *fakeTestFlag {
		if hasFakeCred {
			// Fake test uses the magic phone number.
			tTestClient := twilio.NewSmsClient(sidTest, tokenTest, "+15005550006", to...)
			results, errs := tTestClient.Send("If every phone number has 201 CREATED status then the test passed.")

			if len(errs) > 0 {
				printf(LFakeT, "Failed, errors: %v", errs)
			} else {
				printf(LFakeT, "Successful, result: %v", results)
			}
		} else {
			printf(LFakeT, "Please provide test credentials before running the fake test.")
		}
	}
	// Testing with live credentials.
	tClient := twilio.NewSmsClient(sid, token, from, to...)
	if *liveTestFlag {
		if hasLiveCred {

			results, errs := tClient.Send("RtxNotificator is running! " + strconv.Itoa(generator.Intn(9999)))

			if len(errs) > 0 {
				printf(LLiveT, "Failed, errors: %v", errs)
			} else {
				printf(LLiveT, "Successful, result: %v", results)
			}
		} else {
			printf(LLiveT, "Please provide test credentials before running the fake test.")
		}
	}

	if !hasLiveCred {
		fmt.Printf("[Error] Please supply Twilio SID and Authentiaction token via flags or the environment variables.")
		os.Exit(1)
	}

	printf(LInfo, "RtxNotificator started with parameters: \n\tSender -> %v \n\tRecipients -> %v\n\tPolling Interval -> %v\n", from, to, pollInterval)

	// Start the polling.
	startPolling(tClient, nvidiade.NewStockClient(), pollInterval, apiservice.Rtx3060Ti, apiservice.Rtx3070, apiservice.Rtx3080, apiservice.Rtx3090)
}

func startPolling(msgClient msgservice.SenderClient, apiClient apiservice.GpuStockClient, pollInterval time.Duration, gpus ...apiservice.Gpu) {
	for {
		<-time.After(pollInterval)
		go poll(msgClient, apiClient, gpus...)
	}
}

func poll(msgClient msgservice.SenderClient, stockClient apiservice.GpuStockClient, gpus ...apiservice.Gpu) {
	result, err := stockClient.CheckStock(gpus...)
	if err != nil {
		printf(LError, "Polling error: %v", err)
	}

	for _, s := range result {
		// Send the alert string via the msg client.
		if s.ShouldAlert() {
			msg := s.AlertString()
			res, _ := msgClient.Send(msg)

			printf(LPoll, "Alert sent for %v result: %v \n\tSentContent: %v", s.Gpu, res, msg)
		}
	}

	printf(LPoll, "Result from %v: %v", stockClient.Name(), result)
}

func printf(label, msg string, v ...interface{}) {
	fmt.Printf("[%v][%v] %v\n", label, time.Now().Format("15:04:05"), fmt.Sprintf(msg, v...))
}
