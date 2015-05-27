package main

import (
	"github.com/codegangsta/cli"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ImportMaster struct {
	session         *mgo.Session
	ethereum        *eth.Ethereum
	txCollection    *mgo.Collection
	blockCollection *mgo.Collection
}

type Transaction struct {
	TxHash    string `bson:"tx_hash"`
	Recipient string
	From      string
	Amount    string
	Price     string
	GasLimit  string `bson:"gas_limit"`
	Payload   []byte
	BlockId   *bson.ObjectId `bson:"block_id,omitempty"`
}

func (self *ImportMaster) parseTx(tx *types.Transaction, blockId *bson.ObjectId) *Transaction {
	hash := tx.Hash().Hex()
	from, err := tx.From()
	if err != nil {
		utils.Fatalf("Could not parse from address: %v", err)
	}
	var recipient string
	if tx.Recipient != nil {
		recipient = tx.Recipient.Hex()
	}
	txx := &Transaction{hash, recipient, from.Hex(), tx.Amount.String(), tx.Price.String(), tx.GasLimit.String(), tx.Payload, blockId}
	return txx
}

type Block struct {
	BlockHash   string `bson:"block_hash"`
	ParentHash  string `bson:"parent_hash"`
	UncleHash   string `bson:"uncle_hash"`
	Coinbase    string `bson:"coin_base"`
	Root        string `bson:"root"`
	TxHash      string `bson:"tx_hash"`
	ReceiptHash string `bson:"receipt_hash"`
	Number      string
	Difficulty  string
	GasLimit    string `bson:"gas_limit"`
	GasUsed     string `bson:"gas_used"`
	Time        uint64
	TxAmount    uint64 `bson:"tx_amount"`
	Extra       string `bson:"extra"`
	Nonce       string
	StorageSize string         `bson:"storage_size"`
	MixDigest   string         `bson:"mix_digest"`
	Processed   bool           `bson:"processed"`
	Id          *bson.ObjectId `bson:"_id,omitempty"`
}

func NewImporter(ctx *cli.Context) *ImportMaster {
	importer := new(ImportMaster)
	mongoUrl := ctx.String("mongo-url")
	db := ctx.String("mongo-database")

	glog.V(logger.Info).Infoln("Connecting to MongoDB db '", db, "'using", mongoUrl)

	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		panic(err)
	}
	importer.session = session
	importer.txCollection = session.DB(db).C("transactions")
	importer.blockCollection = session.DB(db).C("blocks")

	return importer
}

func (self *ImportMaster) importBlock(block *types.Block) {
	blockHash := block.Header().Hash().Hex()
	txAmount := uint64(len(block.Transactions()))

	glog.V(logger.Info).Infoln("Importing block", blockHash, "Hash with ", txAmount, "transactions")
	extData := string(block.Header().Extra[:])

	err := self.blockCollection.Insert(&Block{blockHash, block.ParentHash().Hex(), block.Header().UncleHash.Hex(), block.Header().Coinbase.Hex(), block.Header().Root.Hex(), block.Header().TxHash.Hex(), block.Header().ReceiptHash.Hex(), block.Header().Number.String(), block.Header().Difficulty.String(), block.Header().GasLimit.String(), block.Header().GasUsed.String(), block.Header().Time, txAmount, extData, string(block.Nonce()), block.Size().String(), block.Header().MixDigest.Hex(), false, nil})
	if err != nil {
		clilogger.Infoln(err)
	}
	result := Block{}
	err = self.blockCollection.Find(bson.M{"block_hash": blockHash}).One(&result)
	if err != nil {
		utils.Fatalf("Could not find the block we just added, saving faild: %v", err)
	}
	for _, tx := range block.Transactions() {
		self.importTx(tx, result.Id)
	}
}
func (self *ImportMaster) importTx(tx *types.Transaction, blockId *bson.ObjectId) {
	if glog.V(logger.Info) {
		glog.Infoln("Importing tx", tx.Hash().Hex())
	}
	err := self.txCollection.Insert(self.parseTx(tx, blockId))
	if err != nil {
		clilogger.Infoln(err)
	}
}
