package fixture

import (
	"math/big"
	"sync"
	"time"
)

type fileTransferHandler struct{}
type FileTransferHandler struct{}
type transferAlias = **fileTransferHandler

var (
	handler                     fileTransferHandler
	downloadHandler             fileTransferHandler
	cachedfiletransferhandler   fileTransferHandler
	cachedFileTransferHandlers  fileTransferHandler
	TRANSFERHANDLER             fileTransferHandler
	downloadTransferHandler     transferAlias
	downloadHandlers            fileTransferHandler // want: namingclarity: "downloadHandlers" should be named
	downloadhandler             fileTransferHandler // want: namingclarity: "downloadhandler" should be named
	banana                      fileTransferHandler // want: namingclarity: "banana" should be named
	downloadFileTransferHandler FileTransferHandler
	cachedfiletransferhandlerss FileTransferHandler // want: namingclarity: "cachedfiletransferhandlerss" should be named
)

func exportedType() {
	var fileTransferHandler FileTransferHandler
	var handler FileTransferHandler         // want: namingclarity: "handler" should be named
	var downloadHandler FileTransferHandler // want: namingclarity: "downloadHandler" should be named
	_ = fileTransferHandler
	_ = handler
	_ = downloadHandler
}

func resolvedVariables() {
	var payload any = FileTransferHandler{}
	payload = FileTransferHandler{}
	var userId any = FileTransferHandler{} // want: namingclarity: disallowed initialism casing "Id"
	_ = payload
	_ = userId
	for _, handler := range []fileTransferHandler{{}} {
		_ = handler
	}
	for _, downloadHandlers := range []fileTransferHandler{{}} { // want: namingclarity: "downloadHandlers" should be named
		_ = downloadHandlers
	}
}

type Subscription struct{}
type DomainResource struct{}

var (
	subscription                 Subscription
	subscriptionAlias            Subscription
	subscriptions                Subscription
	subscriptionAliases          Subscription
	primaryLink                  Subscription
	backupLinks                  Subscription
	priorityBox                  Subscription
	priorityBoxes                Subscription
	priorityBoxess               Subscription // want: namingclarity: "priorityBoxess" should be
	backupLinkss                 Subscription // want: namingclarity: "backupLinkss" should be
	prioritybox                  Subscription // want: namingclarity: "prioritybox" should be
	SubscriptionAlias            Subscription // want: namingclarity: "SubscriptionAlias" should be
	inventorySubscriptionAliases Subscription // want: namingclarity: "inventorySubscriptionAliases" should be
	subscriptionAliasesS         Subscription // want: namingclarity: "subscriptionAliasesS" should be
	renewalSubscription          Subscription // want: namingclarity: "renewalSubscription" should be
	inventory                    DomainResource
	inventoryCfg                 DomainResource // want: namingclarity: disallowed identifier term "cfg"
)

var (
	now            time.Time
	createdAt      time.Time
	lockedUntil    time.Time
	bananaTime     time.Time
	createdAtValue time.Time // want: namingclarity: "createdAtValue" should be
	createdat      time.Time // want: namingclarity: "createdat" should be
	cachedtime     time.Time // want: namingclarity: "cachedtime" should be
)

func timeName() {
	var banana time.Time        // want: namingclarity: "banana" should be
	var now FileTransferHandler // want: namingclarity: "now" should be
	_ = banana
	_ = now
}

var (
	totalBigInt big.Int
	countInt    int // want: namingclarity: meaningless representation suffix prohibited: countInt
	userID      int
	serverURL   string
	userUUID    string
	userId      int    // want: namingclarity: disallowed initialism casing "Id"
	serverUrl   string // want: namingclarity: disallowed initialism casing "Url"
	userUuid    string // want: namingclarity: disallowed initialism casing "Uuid"
)

func suffixAssignments() {
	totalBigInt := big.Int{}
	countInt := 1 // want: namingclarity: meaningless representation suffix prohibited: countInt
	_ = totalBigInt
	_ = countInt
}

func suffixWithoutMatchingType() {
	var totalBigInt int // want: namingclarity: meaningless representation suffix prohibited: totalBigInt
	_ = totalBigInt
}

var (
	pool                     sync.Pool
	rootCertPool             int
	databasePoolOptions      int
	databasePool             int       // want: namingclarity: disallowed identifier term "pool"
	otherDatabasePoolOptions string    // want: namingclarity: disallowed identifier term "pool"
	workerPool               int       // want: namingclarity: disallowed identifier term "pool"
	poolCfg                  sync.Pool // want: namingclarity: disallowed identifier term "cfg"
	poolId                   sync.Pool // want: namingclarity: disallowed initialism casing "Id"
	poolWrapper              sync.Pool // want: namingclarity: "poolWrapper" should be
	poolTempobj              sync.Pool // want: namingclarity: disallowed compound pattern "tempobj"
	temporaryObject          int
	mytempobject             int // want: namingclarity: disallowed compound pattern "tempobj"
)

var (
	err         error
	readErr     error
	ErrNotFound error
	ErrÉchec    error
	Erré        error // want: namingclarity: "Erré" should be
	readError   error // want: namingclarity: "readError" should be
	ErrnotFound error // want: namingclarity: "ErrnotFound" should be
)

func errorName() {
	var error error // want: namingclarity: "error" should be
	_ = error
}

type Server struct{}

func (server *Server) Serve()                                               {}
func (s *Server) Start()                                                    {} // want: namingclarity: single-letter receiver prohibited: s || namingclarity: "s" should be named
func (wrongReceiver *Server) Stop()                                         {} // want: namingclarity: "wrongReceiver" should be named
func serve(count int)                                                       {}
func rejectParam(x int)                                                     {} // want: namingclarity: single-letter argument prohibited: x
func acceptTypes(fileTransferHandler *FileTransferHandler) (server *Server) { return nil }
func rejectParamType(wrongParameter *FileTransferHandler)                   {}             // want: namingclarity: "wrongParameter" should be named
func rejectResultType() (wrongResult *Server)                               { return nil } // want: namingclarity: "wrongResult" should be named
