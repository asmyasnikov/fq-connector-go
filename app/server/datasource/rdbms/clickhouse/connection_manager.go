package clickhouse

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	api_common "github.com/ydb-platform/fq-connector-go/api/common"
	rdbms_utils "github.com/ydb-platform/fq-connector-go/app/server/datasource/rdbms/utils"
	"github.com/ydb-platform/fq-connector-go/common"
)

var _ rdbms_utils.ConnectionManager = (*connectionManager)(nil)

type connectionManager struct {
	rdbms_utils.ConnectionManagerBase
}

func (c *connectionManager) Make(
	ctx context.Context,
	logger *zap.Logger,
	dsi *api_common.TDataSourceInstance,
) (rdbms_utils.Connection, error) {
	if dsi.GetCredentials().GetBasic() == nil {
		return nil, fmt.Errorf("currently only basic auth is supported")
	}

	switch dsi.Protocol {
	case api_common.EProtocol_NATIVE:
		return makeConnectionNative(ctx, logger, dsi, c.QueryLoggerFactory.Make(logger))
	case api_common.EProtocol_HTTP:
		return makeConnectionHTTP(ctx, logger, dsi, c.QueryLoggerFactory.Make(logger))
	default:
		return nil, fmt.Errorf("can not run ClickHouse connection with protocol '%v'", dsi.Protocol)
	}
}

func (*connectionManager) Release(logger *zap.Logger, conn rdbms_utils.Connection) {
	common.LogCloserError(logger, conn, "close clickhouse connection")
}

func NewConnectionManager(cfg rdbms_utils.ConnectionManagerBase) rdbms_utils.ConnectionManager {
	return &connectionManager{ConnectionManagerBase: cfg}
}
