package handler

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/chaos"
	controllers "github.com/LGU-SE-Internal/chaos-experiment/controllers"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

// JVM Return Value Type
type JVMReturnType int

const (
	StringReturn JVMReturnType = 1
	IntReturn    JVMReturnType = 2
)

// JVM Memory Type
type JVMMemoryType int

const (
	HeapMemory  JVMMemoryType = 1
	StackMemory JVMMemoryType = 2
)

// JVMLatencySpec defines the JVM latency chaos injection parameters
// Updated to use flattened MethodIdx
type JVMLatencySpec struct {
	Duration        int `range:"1-60" description:"Time Unit Minute"`
	System          int `range:"0-0" dynamic:"true" description:"String"`
	MethodIdx       int `range:"0-0" dynamic:"true" description:"Flattened app+method index"`
	LatencyDuration int `range:"1-5000" description:"Latency in ms"`
}

func (s *JVMLatencySpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	methods, err := resourcelookup.GetSystemCache(system).GetAllJVMMethods()
	if err != nil {
		return "", fmt.Errorf("failed to get JVM methods: %w", err)
	}

	if s.MethodIdx < 0 || s.MethodIdx >= len(methods) {
		return "", fmt.Errorf("method index out of range: %d (max: %d)", s.MethodIdx, len(methods)-1)
	}

	methodPair := methods[s.MethodIdx]
	appName := methodPair.AppName
	className := methodPair.ClassName
	methodName := methodPair.MethodName

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	optss := []chaos.OptJVMChaos{
		chaos.WithJVMClass(className),
		chaos.WithJVMMethod(methodName),
		chaos.WithJVMLatencyDuration(s.LatencyDuration),
	}

	return controllers.CreateJVMChaos(cli, ctx, ns, appName,
		chaosmeshv1alpha1.JVMLatencyAction, duration, annotations, labels, optss...)
}

// JVMReturnSpec defines the JVM return value chaos injection parameters
// Updated to use flattened MethodIdx
type JVMReturnSpec struct {
	Duration       int           `range:"1-60" description:"Time Unit Minute"`
	System         int           `range:"0-0" dynamic:"true" description:"Namespace Index (0-based)"`
	MethodIdx      int           `range:"0-0" dynamic:"true" description:"Flattened app+method index"`
	ReturnType     JVMReturnType `range:"1-2" description:"Return Type (1=String, 2=Int)"`
	ReturnValueOpt int           `range:"0-1" description:"Return value option (0=Default, 1=Random)"`
}

func (s *JVMReturnSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	methods, err := resourcelookup.GetSystemCache(system).GetAllJVMMethods()
	if err != nil {
		return "", fmt.Errorf("failed to get JVM methods: %w", err)
	}

	if s.MethodIdx < 0 || s.MethodIdx >= len(methods) {
		return "", fmt.Errorf("method index out of range: %d (max: %d)", s.MethodIdx, len(methods)-1)
	}

	methodPair := methods[s.MethodIdx]
	appName := methodPair.AppName
	className := methodPair.ClassName
	methodName := methodPair.MethodName

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	optss := []chaos.OptJVMChaos{
		chaos.WithJVMClass(className),
		chaos.WithJVMMethod(methodName),
	}

	if s.ReturnValueOpt == 0 {
		// Use default value
		if s.ReturnType == StringReturn {
			optss = append(optss, chaos.WithJVMDefaultStringReturn())
		} else {
			optss = append(optss, chaos.WithJVMDefaultIntReturn())
		}
	} else {
		// Use random value
		if s.ReturnType == StringReturn {
			optss = append(optss, chaos.WithJVMRandomStringReturn(8))
		} else {
			optss = append(optss, chaos.WithJVMRandomIntReturn(1, 1000))
		}
	}

	return controllers.CreateJVMChaos(cli, ctx, ns, appName,
		chaosmeshv1alpha1.JVMReturnAction, duration, annotations, labels, optss...)
}

// JVMExceptionSpec defines the JVM exception injection parameters
// Updated to use flattened MethodIdx
type JVMExceptionSpec struct {
	Duration     int `range:"1-60" description:"Time Unit Minute"`
	System       int `range:"0-0" dynamic:"true" description:"String"`
	MethodIdx    int `range:"0-0" dynamic:"true" description:"Flattened app+method index"`
	ExceptionOpt int `range:"0-1" description:"Exception option (0=Default, 1=Random)"`
}

func (s *JVMExceptionSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	methods, err := resourcelookup.GetSystemCache(system).GetAllJVMMethods()
	if err != nil {
		return "", fmt.Errorf("failed to get JVM methods: %w", err)
	}

	if s.MethodIdx < 0 || s.MethodIdx >= len(methods) {
		return "", fmt.Errorf("method index out of range: %d (max: %d)", s.MethodIdx, len(methods)-1)
	}

	methodPair := methods[s.MethodIdx]
	appName := methodPair.AppName
	className := methodPair.ClassName
	methodName := methodPair.MethodName

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	optss := []chaos.OptJVMChaos{
		chaos.WithJVMClass(className),
		chaos.WithJVMMethod(methodName),
	}

	if s.ExceptionOpt == 0 {
		// Use default exception
		optss = append(optss, chaos.WithJVMDefaultException())
	} else {
		// Use random exception
		randomExceptions := []string{
			"java.io.IOException(\"Random failure\")",
			"java.lang.IllegalArgumentException(\"Invalid argument\")",
			"java.lang.NullPointerException()",
			"java.lang.RuntimeException(\"Unexpected error\")",
			"java.sql.SQLException(\"Database error\")",
		}
		// Pick a random exception from the list
		randomIndex := rand.Intn(len(randomExceptions))
		optss = append(optss, chaos.WithJVMException(randomExceptions[randomIndex]))
	}

	return controllers.CreateJVMChaos(cli, ctx, ns, appName,
		chaosmeshv1alpha1.JVMExceptionAction, duration, annotations, labels, optss...)
}

// JVMGCSpec defines the JVM garbage collector chaos injection parameters
type JVMGCSpec struct {
	Duration int `range:"1-60" description:"Time Unit Minute"`
	System   int `range:"0-0" dynamic:"true" description:"String"`
	AppIdx   int `range:"0-0" dynamic:"true" description:"App Index"`
}

func (s *JVMGCSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	appLabels, err := resourcelookup.GetSystemCache(system).GetAllAppLabels(ctx, ns, defaultAppLabel)
	if err != nil {
		return "", fmt.Errorf("failed to get app labels: %w", err)
	}

	if s.AppIdx < 0 || s.AppIdx >= len(appLabels) {
		return "", fmt.Errorf("app index out of range: %d (max: %d)", s.AppIdx, len(appLabels)-1)
	}

	appName := appLabels[s.AppIdx]
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	return controllers.CreateJVMChaos(cli, ctx, ns, appName,
		chaosmeshv1alpha1.JVMGCAction, duration, annotations, labels)
}

// JVMCPUStressSpec defines the JVM CPU stress chaos injection parameters
// Updated to use flattened MethodIdx
type JVMCPUStressSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	System    int `range:"0-0" dynamic:"true" description:"String"`
	MethodIdx int `range:"0-0" dynamic:"true" description:"Flattened app+method index"`
	CPUCount  int `range:"1-8" description:"Number of CPU cores to stress"`
}

func (s *JVMCPUStressSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	methods, err := resourcelookup.GetSystemCache(system).GetAllJVMMethods()
	if err != nil {
		return "", fmt.Errorf("failed to get JVM methods: %w", err)
	}

	if s.MethodIdx < 0 || s.MethodIdx >= len(methods) {
		return "", fmt.Errorf("method index out of range: %d (max: %d)", s.MethodIdx, len(methods)-1)
	}

	methodPair := methods[s.MethodIdx]
	appName := methodPair.AppName
	className := methodPair.ClassName
	methodName := methodPair.MethodName

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	optss := []chaos.OptJVMChaos{
		chaos.WithJVMClass(className),
		chaos.WithJVMMethod(methodName),
		chaos.WithJVMStressCPUCount(s.CPUCount),
	}

	return controllers.CreateJVMChaos(cli, ctx, ns, appName,
		chaosmeshv1alpha1.JVMStressAction, duration, annotations, labels, optss...)
}

// JVMMemoryStressSpec defines the JVM memory stress chaos injection parameters
// Updated to use flattened MethodIdx
type JVMMemoryStressSpec struct {
	Duration  int           `range:"1-60" description:"Time Unit Minute"`
	System    int           `range:"0-0" dynamic:"true" description:"Namespace Index (0-based)"`
	MethodIdx int           `range:"0-0" dynamic:"true" description:"Flattened app+method index"`
	MemType   JVMMemoryType `range:"1-2" description:"Memory Type (1=Heap, 2=Stack)"`
}

func (s *JVMMemoryStressSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	methods, err := resourcelookup.GetSystemCache(system).GetAllJVMMethods()
	if err != nil {
		return "", fmt.Errorf("failed to get JVM methods: %w", err)
	}

	if s.MethodIdx < 0 || s.MethodIdx >= len(methods) {
		return "", fmt.Errorf("method index out of range: %d (max: %d)", s.MethodIdx, len(methods)-1)
	}

	methodPair := methods[s.MethodIdx]
	appName := methodPair.AppName
	className := methodPair.ClassName
	methodName := methodPair.MethodName

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	// Convert memory type
	memType := "heap"
	if s.MemType == StackMemory {
		memType = "stack"
	}

	optss := []chaos.OptJVMChaos{
		chaos.WithJVMClass(className),
		chaos.WithJVMMethod(methodName),
		chaos.WithJVMStressMemType(memType),
	}

	return controllers.CreateJVMChaos(cli, ctx, ns, appName,
		chaosmeshv1alpha1.JVMStressAction, duration, annotations, labels, optss...)
}

// SQL types for JVMMySQL
type MySQLType int

const (
	AllSQL     MySQLType = 0
	SelectSQL  MySQLType = 1
	InsertSQL  MySQLType = 2
	UpdateSQL  MySQLType = 3
	DeleteSQL  MySQLType = 4
	ReplaceSQL MySQLType = 5
)

// MySQL connector versions
type MySQLConnectorVersion int

const (
	MySQL5 MySQLConnectorVersion = 5
	MySQL8 MySQLConnectorVersion = 8
)

// JVMMySQLLatencySpec defines the JVM MySQL latency chaos injection parameters
type JVMMySQLLatencySpec struct {
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	System      int `range:"0-0" dynamic:"true" description:"String"`
	DatabaseIdx int `range:"0-0" dynamic:"true" description:"Flattened app+database+table index"`
	LatencyMs   int `range:"10-5000" description:"Latency in ms"`
}

func (s *JVMMySQLLatencySpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	dbOps, err := resourcelookup.GetSystemCache(system).GetAllDatabaseOperations()
	if err != nil {
		return "", fmt.Errorf("failed to get database operations: %w", err)
	}

	if s.DatabaseIdx < 0 || s.DatabaseIdx >= len(dbOps) {
		return "", fmt.Errorf("database operation index out of range: %d (max: %d)", s.DatabaseIdx, len(dbOps)-1)
	}

	dbOp := dbOps[s.DatabaseIdx]
	appName := dbOp.AppName
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	// Convert the operation type to lowercase for Chaos Mesh
	sqlTypeStr := strings.ToLower(dbOp.OperationType)

	optss := []chaos.OptJVMChaos{
		chaos.WithJVMMySQLConnector("8"), // Hardcoded to version 8
		chaos.WithJVMMySQLDatabase(dbOp.DBName),
		chaos.WithJVMMySQLTable(dbOp.TableName),
		chaos.WithJVMMySQLType(sqlTypeStr),
		chaos.WithJVMLatencyDuration(s.LatencyMs),
	}

	return controllers.CreateJVMChaos(cli, ctx, ns, appName,
		chaosmeshv1alpha1.JVMMySQLAction, duration, annotations, labels, optss...)
}

// JVMMySQLExceptionSpec defines the JVM MySQL exception chaos injection parameters
type JVMMySQLExceptionSpec struct {
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	System      int `range:"0-0" dynamic:"true" description:"String"`
	DatabaseIdx int `range:"0-0" dynamic:"true" description:"Flattened app+database+table index"`
}

func (s *JVMMySQLExceptionSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annotations != nil {
		annotations = conf.Annotations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace
	system := conf.System

	dbOps, err := resourcelookup.GetSystemCache(system).GetAllDatabaseOperations()
	if err != nil {
		return "", fmt.Errorf("failed to get database operations: %w", err)
	}

	if s.DatabaseIdx < 0 || s.DatabaseIdx >= len(dbOps) {
		return "", fmt.Errorf("database operation index out of range: %d (max: %d)", s.DatabaseIdx, len(dbOps)-1)
	}

	dbOp := dbOps[s.DatabaseIdx]
	appName := dbOp.AppName
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	// Convert the operation type to lowercase for Chaos Mesh
	sqlTypeStr := strings.ToLower(dbOp.OperationType)

	// Always use "BOOM" as the exception message
	exceptionMsg := "BOOM"

	optss := []chaos.OptJVMChaos{
		chaos.WithJVMMySQLConnector("8"), // Hardcoded to version 8
		chaos.WithJVMMySQLDatabase(dbOp.DBName),
		chaos.WithJVMMySQLTable(dbOp.TableName),
		chaos.WithJVMMySQLType(sqlTypeStr),
		chaos.WithJVMException(exceptionMsg),
	}

	return controllers.CreateJVMChaos(cli, ctx, ns, appName,
		chaosmeshv1alpha1.JVMMySQLAction, duration, annotations, labels, optss...)
}
