// Package rum implements the continous flow of concurrent funcs manager which can be regisreted via server can be called via clinet
package rum

// flow:
//
//												                Create Profile:
//											                            ↓
//											                    "profile" includes:
//											         kit: provides the descirption of the profile that includes:
//											             "services": contains descirption of the service which include:
//										                             "time-format" ->  deactivate the service, remove the service,activation time, retry call if the dispatch failed, on what time to invoke, duration of delay per dispatch,
//									                                 "dispatcher" -> controls the number of registered funcs
//								                                     "budget"     -> controls the mode budget
//													     "time-format": same as services but for the profile
//							                                  Call The Profile:
//							                                            ↓
//					                                        grpc accepts the call & triggers hub through channel
//				                                                     ↓
//			                                               hub performs as per the call:
//						                                 onPost-> fetches the service -> reads the format -> performs the write -> waits for the monitor -> write publishes the work -> tickFetch fetches the result -> Paper publishes the result -> result is passed to the client
//	                                                  onDeactivate-> read desc ->  temporarly remove the service or profile
//	                                                  onActivate-> read desc ->  find the deactivate serivce or profile -> activates the service or profile
//	                                                  onRemove-> read desc ->  remove the service or profile
