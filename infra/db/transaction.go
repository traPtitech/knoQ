package db

import (
	"context"

	"gorm.io/gorm"
)

type gormTransactionManager struct {
	db *gorm.DB
}

// context用のキー
type txKey struct{}

func NewTransactionManager(db *gorm.DB) *gormTransactionManager {
	return &gormTransactionManager{db: db}
}

// Service以上でtransactionを張りたい場合
func (repo *gormRepository) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// トランザクション(tx)をcontextにセットして実行
		ctxWithTx := context.WithValue(ctx, txKey{}, tx)
		return fn(ctxWithTx)
	})
}

// transactionを張らなくてもデフォルトのdbで実行できる
func getTx(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}
