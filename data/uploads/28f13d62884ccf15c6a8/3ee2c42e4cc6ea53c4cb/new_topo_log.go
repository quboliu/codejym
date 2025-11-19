
  cnSQLVerificationTimeout  = 30 * time.Second
  cnSQLQueryTimeout         = 30 * time.Second
  cnSQLVerificationInterval = 200 * time.Millisecond

  cnHoldonReleaseBackoffBase = 1 * time.Second
  cnHoldonReleaseBackoffCap  = 5 * time.Second




  defer func() {
    if releaseHoldon {
      conns = t.ensureHoldonReleased(conns)
    }
    t.closeCNs(conns)
  }()


func (t *TopologyAggregator) execSetHoldon(conns []*pgx.Conn, on bool) error {
  val := "false"
  if on {
    val = "true"
  }
  sql := fmt.Sprintf("set hold on %s", val)
  EtcdLogger.Info("updating CN holdon state via SQL",
    zap.Bool("on", on),
    zap.Int("cn_count", len(conns)))
  for idx, c := range conns {
    ctx, cancel := context.WithTimeout(context.Background(), cnSQLQueryTimeout)
    _, err := c.Exec(ctx, sql)
    cancel()
    if err != nil {
      return fmt.Errorf("set hold on %s failed on connection %d: %w", val, idx, err)
    }
  }
  EtcdLogger.Info("CN holdon state updated successfully",
    zap.Bool("on", on),
    zap.Int("cn_count", len(conns)))
  return nil
}

// ensureHoldonReleased makes sure holdon is eventually turned off.
// It retries with exponential backoff and reconnects if sessions dropped.
// Returns the latest slice of connections so the caller can close them.
func (t *TopologyAggregator) ensureHoldonReleased(conns []*pgx.Conn) []*pgx.Conn {
  current := conns
  backoff := cnHoldonReleaseBackoffBase
  attempt := 0

  for {
    attempt++
    if len(current) == 0 {
      EtcdLogger.Warn("no active CN connections available for holdon release, attempting reconnect",
        zap.Int("attempt", attempt))
      var err error
      current, err = t.connectCNs()
      if err != nil {
        EtcdLogger.Warn("reconnect CN postgres failed during holdon release",
          zap.Int("attempt", attempt),
          zap.Error(err))
        time.Sleep(backoff)
        if backoff < cnHoldonReleaseBackoffCap {
          backoff *= 2
          if backoff > cnHoldonReleaseBackoffCap {
            backoff = cnHoldonReleaseBackoffCap
          }
        }
        continue
      }
      EtcdLogger.Info("reconnected CN sessions for holdon release",
        zap.Int("attempt", attempt),
        zap.Int("cn_count", len(current)))
    }

    if err := t.execSetHoldon(current, false); err != nil {
      EtcdLogger.Warn("set holdon off attempt failed",
        zap.Int("attempt", attempt),
        zap.Error(err))
      t.closeCNs(current)
      current = nil
      time.Sleep(backoff)
      if backoff < cnHoldonReleaseBackoffCap {
        backoff *= 2
        if backoff > cnHoldonReleaseBackoffCap {
          backoff = cnHoldonReleaseBackoffCap
        }
      }
      continue
    }

    EtcdLogger.Info("holdon release succeeded",
      zap.Int("attempt", attempt))
    return current
  }
}



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
      var gotStr string
      err := c.QueryRow(ctx, "show cluster_topo_version").Scan(&gotStr)
      cancel()
      if err != nil {
        EtcdLogger.Warn("fetch CN topology version via SQL failed",
          zap.Int("conn_index", idx),
          zap.Error(err))
        allOK = false
        break
      }
      // Parse SHOW result
      gotStr = strings.TrimSpace(gotStr)
      got, parseErr := strconv.ParseInt(gotStr, 10, 64)
      if parseErr != nil {
        EtcdLogger.Warn("parse CN topology version via SQL failed",
          zap.Int("conn_index", idx),
          zap.String("raw_version", gotStr),
          zap.Error(parseErr))
        allOK = false
        break
      }
      if got < want {
        allOK = false
        now := time.Now()
        if now.After(nextLog) {
          EtcdLogger.Info("CN topology version behind target",
            zap.Int("conn_index", idx),
            zap.Int64("current_version", got),
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


        if releaseHoldon {
          conns = t.ensureHoldonReleased(conns)
        }
