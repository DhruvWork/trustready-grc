import Foundation
import TrustReadyAgentShared
import os

private let log = Logger(subsystem: "com.trustready.agent.helper", category: "main")

let helper = Helper()
let listener = NSXPCListener(machServiceName: TrustReadyAgentHelperConstants.machServiceName)
listener.delegate = helper
listener.resume()
log.info("listening on \(TrustReadyAgentHelperConstants.machServiceName, privacy: .public)")

RunLoop.main.run()
