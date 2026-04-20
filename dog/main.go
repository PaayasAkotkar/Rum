package dog

// flow:
// ------------------------------------------------
//
// gemini-proposed-scenario:
// Imagine we are deploying RUM in a production
// environment where we have 50 different
// registered functions (models/agents).
// One specific model—let's call
// it the DeepAnalyzer—is very resource-heavy
// and takes about 5 seconds to complete,
// while the others take 50ms.
// end gemini-proposed-scenario
//
// ------------------------------------------------
// ------------------------------------------------
// the gemini's proposed is kind of logical for
// my rum-model
// so I created this package to solve the issuse
// and answer the gemini proposed question.
// ------------------------------------------------
// ------------------------------------------------
// what I noted:
// 1.everythings needs to be clean and easy for
//   the client to manage.
// 2.every registered func in that profile
//   needs to be monitor than simply monitor one
//   single profile.
//   meaning that suppose 50 calls made
//   -> we basically
//   ticking the funcs in and removing
//   the duplicate calls
// for example you can checkout
//  rumdog.ExampleMultiplePoliciesConcurrent()
// 3. settings-> this made things for clean
//    and easy to maintain for misc things
//    that I was having doubt the clinet
//    may set this that....
// ------------------------------------------------
// ------------------------------------------------
// guide:
// 1. create the rumdog
// 2. keep the dog.Watch() in the main func
// 3. create pollicies and regisreted it
//    note: make sure that you put the time.sleep
//          in both regisreted func and
//          outisde so that calculation
//          can be sync
// 4. you can create a waitgroups to fetch live
// results of these policis or use something
// like uber.fx for better
// 5. and that's it you get the results
// ------------------------------------------------
// ------------------------------------------------
// flow:
// $core:= [process<setTimeout]
// Keep the Hub open
//               V
// [we'll only focus on core part]
// monitor create writes start report
// and requests ticking
//               V
// ticking begans in tickSinglePolicy
// for each req it checks $core
//               V
// when the client's func is done ->
// processDone writes
// the report and -> checks $core
// & cleans the policy funcs
// for each process-done the pubsub is activated
// and the report is send to Pakkun
// ------------------------------------------------
// -x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-
