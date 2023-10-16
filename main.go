package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"example.com/kzg-demo/utils"

	"example.com/kzg-demo/types"
	"github.com/protolambda/go-kzg/bls"
)

const PipesNum = 9 // do not change this value

var inputs []*bufio.Scanner
var _inputFiles []*os.File
var outputs []*bufio.Writer
var _outputFiles []*os.File
var _startTime time.Time = time.Now()

func handleSignal() {
	// Create a channel to receive signals
	sigChan := make(chan os.Signal, 1)
	// Notify the channel for SIGINT and SIGTERM signals
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	// Start a separate goroutine to handle the signal
	go func() {
		// Block until a signal is received
		sig := <-sigChan
		fmt.Println("\nReceived signal:", sig)

		// Perform any cleanup or handle the signal as needed
		onExit()

		// Exit the program gracefully
		os.Exit(1)
	}()
}

func onExit() {
	closeIO()
}

func onStart() {
	handleSignal()
	prepareIO()
}

func closeIO() {
	for _, file := range _inputFiles {
		_ = file.Close()
	}
	for _, file := range _outputFiles {
		_ = file.Close()
	}
}

func prepareIO() {
	for i := 0; i < PipesNum; i++ {
		fmt.Printf("Opening %v.fifo.in for reading\n", i)
		file, err := os.OpenFile(fmt.Sprintf("./run/%v.fifo.in", i), os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			log.Fatal(err)
		}

		_inputFiles = append(_inputFiles, file)

		scanner := bufio.NewScanner(file)
		// reader := bufio.NewReader(file)
		inputs = append(inputs, scanner)
	}

	// Open FIFO files for writing
	for i := 0; i < PipesNum; i++ {
		fmt.Printf("Opening %v.fifo.out for writing\n", i)

		file, err := os.OpenFile(fmt.Sprintf("./run/%v.fifo.out", i), os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			log.Fatal(err)
		}

		_outputFiles = append(_outputFiles, file)

		writer := bufio.NewWriter(file)
		outputs = append(outputs, writer)
	}
}

func Output(pipeId int, text string) {
	OutputRaw(pipeId, text)
	// uptime := time.Since(_startTime)
	// OutputRaw(pipeId, fmt.Sprintf("[%v]%v", uptime, text))
}

func OutputRaw(pipeId int, text string) {
	_, err := outputs[pipeId].WriteString(text)
	if err != nil {
		panic(err)
	}
	err = outputs[pipeId].Flush()
	if err != nil {
		panic(err)
	}
}

func demo() error {
	// read number of nodes
	Output(0, "Input number of nodes (from 3 to 8): ")

	inputs[0].Scan()
	nodesCountStr := inputs[0].Text()
	nodesCountStr = strings.Trim(nodesCountStr, "\x00") // workaround
	nodesCountStr = strings.TrimSpace(nodesCountStr)
	nodesCount, err := strconv.Atoi(nodesCountStr)
	if err != nil {
		return err
	}
	if nodesCount < 3 || nodesCount > 8 {
		return errors.New("number of nodes should be in range from 3 to 8")
	}

	controller := types.Controller{}
	nodes := make([]types.Node, nodesCount)

	// input & parse the route
	route := make([]int, 0)
	{

		routeExample := "ABCDEFGH"[:nodesCount]

		Output(0, fmt.Sprintf("Input the route, e.g., %v: ", routeExample))

		inputs[0].Scan()
		routeStr := inputs[0].Text()
		routeStr = strings.Trim(routeStr, "\x00") // workaround
		routeStr = strings.TrimSpace(routeStr)
		if len(routeStr) != nodesCount {
			return fmt.Errorf("unexpected route. %v letters are expected, got %v", nodesCount, len(routeStr))
		}

		// parse the route
		routeUsed := make(map[int]bool)
		for _, ch := range routeStr {
			nodeId := int(ch - 65)
			if nodeId < 0 || nodeId >= nodesCount {
				return fmt.Errorf("unexpected route. Unexpected character %v", string(rune(ch)))
			}
			if _, contains := routeUsed[nodeId]; contains {
				return fmt.Errorf("unexpected route. Duplicated character %v", string(rune(ch)))
			}
			route = append(route, nodeId)
			routeUsed[nodeId] = true
		}
	}

	// show choices for preparing random data
	nodesPrivateData := make([]uint32, 0)
	Output(0, fmt.Sprintf("A. Let controller to generate random data for each node\n"))
	Output(0, fmt.Sprintf("B. Manually input random data for each node\n"))
	Output(0, fmt.Sprintf("Input the choice of random data generation policy, e.g., A: \n"))
	inputs[0].Scan()
	choiceStr := inputs[0].Text()
	choiceStr = strings.Trim(choiceStr, "\x00")
	choiceStr = strings.TrimSpace(choiceStr)
	if choiceStr == "A" {
		for i := 0; i < nodesCount; i++ {
			nodeName := string(rune(65 + i))

			data := rand.Int31()

			secretFr := new(bls.Fr)
			bls.AsFr(secretFr, uint64(data))
			nodes[i].SecretFr = secretFr
			nodes[i].Secret = uint64(data)
			nodesPrivateData = append(nodesPrivateData, uint32(data))

			Output(i+1, fmt.Sprintf("[%v] Private data: %v\n", nodeName, data))
		}
	} else if choiceStr == "B" {
		// read private content of each node

		for i := 0; i < nodesCount; i++ {
			nodeName := string(rune(65 + i))
			Output(0, fmt.Sprintf("Input node %v's private data (uint32 for demo): ", nodeName))

			inputs[0].Scan()
			dataStr := inputs[0].Text()

			dataStr = strings.Trim(dataStr, "\x00")
			dataStr = strings.TrimSpace(dataStr)

			data, err := strconv.Atoi(dataStr)
			if err != nil {
				return err
			}

			secretFr := new(bls.Fr)
			bls.AsFr(secretFr, uint64(data))
			nodes[i].SecretFr = secretFr
			nodes[i].Secret = uint64(data)
			nodesPrivateData = append(nodesPrivateData, uint32(data))

			Output(i+1, fmt.Sprintf("[%v] Private data: %v\n", nodeName, data))
		}
	} else {
		return fmt.Errorf("invalid choice. A or B expected, got %v", choiceStr)
	}

	if !utils.IsUnique(nodesPrivateData) {
		return fmt.Errorf("invalid private data. Private data must be unique for each node")
	}

	Output(0, "Demo begins.\n")

	// controller: initialize polynomial
	procedureSetupBeginTime := time.Now()

	_, _, err = controller.Setup(nodesPrivateData)
	if err != nil {
		return err
	}
	Output(0, "[0] KZG setup completed. Polynomial generated: \n")

	for k, coeff := range controller.Polynomial {
		Output(0, fmt.Sprintf("Polynomial-%v: %v\n", k, coeff.String()))
	}

	Output(0, "[0] KZG setup parameter -- SecretG1: \n")
	for i, g1 := range controller.KzgSettings.SecretG1 {
		Output(0, fmt.Sprintf("SecretG1-%v: \n", i))
		Output(0, fmt.Sprintf("%v: \n", g1.String()))
	}

	Output(0, "[0] KZG setup parameter -- SecretG2: \n")
	for i, g2 := range controller.KzgSettings.SecretG2 {
		Output(0, fmt.Sprintf("SecretG2-%v: \n", i))
		Output(0, fmt.Sprintf("%v: \n", g2.String()))
	}

	for i := 0; i < nodesCount; i++ {
		nodeName := string(rune(65 + i))
		nodes[i].Polynomial = controller.Polynomial

		Output(i+1, fmt.Sprintf("[%v] Parameters received\n", nodeName))
	}

	public := types.PublicStorage{
		KzgSettings:          controller.KzgSettings,
		PolynomialCommitment: controller.Commit(),
	}

	// controller: commit polynomial
	Output(0, fmt.Sprintf("[0] Commit polynomial:\n%v\n", bls.StrG1(public.PolynomialCommitment)))

	procedureSetupInterval := time.Since(procedureSetupBeginTime)
	Output(0, fmt.Sprintf("Setup time cost: %v ms\n", procedureSetupInterval.Milliseconds()))

	Output(0, fmt.Sprintf("Press Enter key to send packets...\n"))
	inputs[0].Scan()

	// input & parse the route
	realRoute := make([]int, 0)
	{

		routeExample := "ABCDEFGH"[:nodesCount]

		Output(0, fmt.Sprintf("Input the real route, e.g., %v: ", routeExample))

		inputs[0].Scan()
		routeStr := inputs[0].Text()
		routeStr = strings.Trim(routeStr, "\x00") // workaround
		routeStr = strings.TrimSpace(routeStr)
		//if len(routeStr) != nodesCount {
		//	return fmt.Errorf("unexpected route. %v letters are expected, got %v", nodesCount, len(routeStr))
		//}

		// parse the route
		//routeUsed := make(map[int]bool)
		for _, ch := range routeStr {
			nodeId := int(ch - 65)

			// at least route should be a letter
			if nodeId < 0 || nodeId >= 26 {
				return fmt.Errorf("unexpected route. Unexpected character %v", string(rune(ch)))
			}

			//if nodeId < 0 || nodeId >= nodesCount {
			//return fmt.Errorf("unexpected route. Unexpected character %v", string(rune(ch)))
			//}
			//if _, contains := routeUsed[nodeId]; contains {
			//	return fmt.Errorf("unexpected route. Duplicated character %v", string(rune(ch)))
			//}
			realRoute = append(realRoute, nodeId)
			//routeUsed[nodeId] = true
		}
	}

	// compute node ID
	routeConsoleID := make(map[int]int)
	{
		realRouteCopy := make([]int, len(realRoute))
		copy(realRouteCopy, realRoute)
		realRouteCopy = append(realRouteCopy, route...)
		sort.Ints(realRouteCopy)

		// assign console ID
		for _, nodeID := range realRouteCopy {
			if _, contains := routeConsoleID[nodeID]; !contains {
				routeConsoleID[nodeID] = len(routeConsoleID)%(PipesNum-1) + 1
			}
		}
	}

	// prepare up to 26 nodes
	for i := nodesCount; i < 26; i++ {
		data := rand.Int31()

		secretFr := new(bls.Fr)
		bls.AsFr(secretFr, uint64(data))

		node := types.Node{}
		node.SecretFr = secretFr
		node.Secret = uint64(data)
		node.Polynomial = controller.Polynomial

		nodes = append(nodes, node)
	}

	// verify the previous node's proof and generate my proof
	var lastProof *bls.G1Point = nil
	var lastNodeName = "0"
	var lastNodeSecret *bls.Fr = nil

	for j := 0; j < len(realRoute); j++ {
		thisNodeID := realRoute[j]
		thisNodeName := string(rune(65 + thisNodeID))

		thisNodeConsoleID := routeConsoleID[thisNodeID]

		nextNodeID := -1
		nextNodeName := "0"
		if j < len(realRoute)-1 {
			nextNodeID = realRoute[j+1]
			nextNodeName = string(rune(65 + nextNodeID))
		}
		Output(0, fmt.Sprintf("Route: %v->%v...\n", thisNodeName, nextNodeName))

		// verifying last node's proof
		if j != 0 {
			Output(thisNodeConsoleID, fmt.Sprintf("[%v] Received and verifying %v's proof: \n", thisNodeName, lastNodeName))
			Output(thisNodeConsoleID, fmt.Sprintf("x=%v, y=%v\n", lastNodeSecret.String(), j-1+1))
			Output(thisNodeConsoleID, fmt.Sprintf("proof:\n%v\n", lastProof.String()))

			var proofVerified bool
			REPEAT_COUNT := 100

			procedureVerifyStartTime := time.Now()
			for k := 0; k < REPEAT_COUNT; k++ {
				y := new(bls.Fr)
				bls.AsFr(y, uint64(j-1+1))

				proofVerified = public.KzgSettings.CheckProofSingle(
					public.PolynomialCommitment, lastProof, lastNodeSecret, y)
			}
			procedureVerifyInterval := float64(time.Since(procedureVerifyStartTime).Nanoseconds()) / float64(1000000) / float64(REPEAT_COUNT)

			if proofVerified {
				Output(thisNodeConsoleID, fmt.Sprintf("[%v] Verification result: \033[0;32m%v\033[0m\n", thisNodeName, proofVerified))
			} else {
				Output(thisNodeConsoleID, fmt.Sprintf("[%v] Verification result: \033[0;31m%v\033[0m\n", thisNodeName, proofVerified))
			}
			Output(thisNodeConsoleID, fmt.Sprintf("[%v] Verification time cost: %.2f ms\n", thisNodeName, procedureVerifyInterval))
		}

		// generating my proof
		{
			Output(thisNodeConsoleID, fmt.Sprintf("[%v] Generating my proof: \n", thisNodeName))
			Output(thisNodeConsoleID, fmt.Sprintf("x=%v, y=%v\n", nodes[thisNodeID].Secret, j+1))

			REPEAT_COUNT := 100

			procedureProveStartTime := time.Now()
			for k := 0; k < REPEAT_COUNT; k++ {
				proof := public.KzgSettings.ComputeProofSingle(nodes[thisNodeID].Polynomial, nodes[thisNodeID].Secret)
				lastProof = proof
				lastNodeName = thisNodeName
				lastNodeSecret = nodes[thisNodeID].SecretFr
			}
			procedureProveInterval := float64(time.Since(procedureProveStartTime).Nanoseconds()) / float64(1000000) / float64(REPEAT_COUNT)

			Output(thisNodeConsoleID, fmt.Sprintf("proof:\n%v\n", lastProof.String()))

			Output(thisNodeConsoleID, fmt.Sprintf("[%v] Prove time cost: %.2f ms\n", thisNodeName, procedureProveInterval))
		}

		// self-verifying my proof
		{
			Output(thisNodeConsoleID, fmt.Sprintf("[%v] Self-verifying %v's proof with y=%v\n", thisNodeName, lastNodeName, j+1))

			var proofVerified bool
			REPEAT_COUNT := 100

			procedureVerifyStartTime := time.Now()
			for k := 0; k < REPEAT_COUNT; k++ {
				y := new(bls.Fr)
				bls.AsFr(y, uint64(j+1))

				proofVerified = public.KzgSettings.CheckProofSingle(
					public.PolynomialCommitment, lastProof, lastNodeSecret, y)
			}
			procedureVerifyInterval := float64(time.Since(procedureVerifyStartTime).Nanoseconds()) / float64(1000000) / float64(REPEAT_COUNT)

			if proofVerified {
				Output(thisNodeConsoleID, fmt.Sprintf("[%v] Verification result: \033[0;32m%v\033[0m\n", thisNodeName, proofVerified))
			} else {
				Output(thisNodeConsoleID, fmt.Sprintf("[%v] Verification result: \033[0;31m%v\033[0m\n", thisNodeName, proofVerified))
			}
			Output(thisNodeConsoleID, fmt.Sprintf("[%v] Verification time cost: %.2f ms\n", thisNodeName, procedureVerifyInterval))
		}

		//Output(0, fmt.Sprintf("Will continue shortly...\n"))
		time.Sleep(500 * time.Millisecond)
	}

	Output(0, "Demo ends.\n")

	return nil
}

func main() {
	onStart()

	for {
		runtime.GOMAXPROCS(1)
		debug.SetGCPercent(-1)
		runtime.GC()

		err := demo()
		if err != nil {
			Output(0, fmt.Sprintf("%v\n", err.Error()))
		}

		Output(0, "Press Enter key to restart this demo...")
		inputs[0].Scan()

		// clear screen
		for i, _ := range outputs {
			//for j := 0; j < 100; j++ {
			//	OutputRaw(i, "\n")
			//}
			OutputRaw(i, "\033[2J\033[;H")
		}
	}
}
