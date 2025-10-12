package main

// func getTables(config *Config, data *database.Conn) {
// 	var err error
// 	var sTables []database.Table

// 	if config.TableName != "" {
// 		sTables = []database.Table{{Name: config.TableName}}
// 	} else {
// 		sTables, err = data.GetTables(data.SSchema, "BASE TABLE", config.Timeout)
// 		ec.CheckErr(err)
// 	}

// 	var tbls []string
// 	for _, t := range sTables {
// 		if config.FilterDef == "" || !config.Filter.MatchString(t.Name) {
// 			tbls = append(tbls, t.Name)
// 		}
// 	}

// 	cTable := config.Table
// 	cLink := config.Link
// 	cView := config.View
// 	cRoutine := config.Routine

// 	// config.Table = true
// 	config.View = false
// 	// config.Link = false
// 	// config.Routine = false

// 	if len(tbls) > 0 {
// 		backupTasker(config, data, tbls)
// 	}

// 	config.Table = cTable
// 	config.Link = cLink
// 	config.View = cView
// 	config.Routine = cRoutine
// }

// func getViews(config *Config, data *database.Conn) {
// 	var err error
// 	var sViews []database.ViewList

// 	if config.ViewName != "" {
// 		sViews = []database.ViewList{{Name: config.ViewName}}
// 	} else {
// 		sViews, err = data.GetViews(data.SSchema, config.Timeout)
// 		ec.CheckErr(err)
// 	}

// 	var views []string
// 	for _, t := range sViews {
// 		if config.FilterDef == "" || !config.Filter.MatchString(t.Name) {
// 			views = append(views, t.Name)
// 		}
// 	}

// 	cTable := config.Table
// 	cLink := config.Link
// 	cView := config.View
// 	cRoutine := config.Routine

// 	config.Table = false
// 	config.View = true
// 	config.Link = false
// 	config.Routine = false

// 	if len(views) > 0 {
// 		backupTasker(config, data, views)
// 	}

// 	config.Table = cTable
// 	config.Link = cLink
// 	config.View = cView
// 	config.Routine = cRoutine
// }

// func getRoutines(config *Config, data *database.Conn) {
// 	var err error
// 	var sRoutines []database.RoutineList

// 	if config.RoutineName != "" {
// 		sRoutines = []database.RoutineList{{Name: config.RoutineName}}
// 	} else {
// 		sRoutines, err = data.GetRoutines(data.SSchema, config.Timeout)
// 		ec.CheckErr(err)
// 	}

// 	var routines []string
// 	for _, t := range sRoutines {
// 		if config.FilterDef == "" || !config.Filter.MatchString(t.Name) {
// 			routines = append(routines, t.Name)
// 		}
// 	}

// 	cTable := config.Table
// 	cLink := config.Link
// 	cView := config.View
// 	cRoutine := config.Routine

// 	config.Table = false
// 	config.View = false
// 	config.Link = false
// 	config.Routine = true

// 	if len(routines) > 0 {
// 		backupTasker(config, data, routines)
// 	}

// 	config.Table = cTable
// 	config.Link = cLink
// 	config.View = cView
// 	config.Routine = cRoutine
// }

// func getIndexes(config *Config, data *database.Conn) {
// 	var err error
// 	var sIndexes []database.IndexList

// 	if config.RoutineName != "" {
// 		sIndexes = []database.IndexList{{Name: config.IndexName}}
// 	} else {
// 		sIndexes, err = data.GetIndexes(data.SSchema, config.Timeout)
// 		ec.CheckErr(err)
// 	}

// 	var indexes []string
// 	for _, t := range sIndexes {
// 		if config.FilterDef == "" || !config.Filter.MatchString(t.Name) {
// 			indexes = append(indexes, t.Name)
// 		}
// 	}

// 	if len(indexes) > 0 {
// 		backupTasker(config, data, indexes)
// 	}
// }
