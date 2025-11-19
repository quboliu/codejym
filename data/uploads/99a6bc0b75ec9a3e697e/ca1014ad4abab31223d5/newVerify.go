func (t *TopologyAggregator) verifyCNTopoSyncViaSQL(conns []*pgx.Conn, want int64) error {
	start := time.Now()
	deadline := start.Add(cnSQLVerificationTimeout)
	nextLog := time.Now()

	EtcdLogger.Info("verifying CN topology version via SQL",
		zap.Int("cn_count", len(conns)),
		zap.Int64("target_version", want))

	for {
		allOK := true
		for idx, c := range conns {
			ctx, cancel := context.WithTimeout(context.Background(), cnSQLQueryTimeout)
			var got sql.NullInt64
			err := c.QueryRow(ctx, "select coalesce(max(version), 0) from he3_cluster_info").Scan(&got)
			cancel()
			if err != nil {
				EtcdLogger.Warn("fetch CN topology version via SQL failed",
					zap.Int("conn_index", idx),
					zap.Error(err))
				allOK = false
				break
			}
			if !got.Valid {
				EtcdLogger.Warn("CN topology version query returned NULL",
					zap.Int("conn_index", idx))
				allOK = false
				break
			}
			currentVersion := got.Int64
			if currentVersion < want {
				allOK = false
				now := time.Now()
				if now.After(nextLog) {
					EtcdLogger.Info("CN topology version behind target",
						zap.Int("conn_index", idx),
						zap.Int64("current_version", currentVersion),
						zap.Int64("target_version", want),
						zap.Duration("elapsed", now.Sub(start)))
					nextLog = now.Add(time.Second)
				}
				break
			}
		}
		if allOK {
			EtcdLogger.Info("CN topology version verified via SQL",
				zap.Int64("target_version", want),
				zap.Duration("elapsed", time.Since(start)))
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting cn sql version >= %d", want)
		}
		time.Sleep(cnSQLVerificationInterval)
	}
}