import Foundation

public enum TrustReadyAgentHelperConstants {
    public static let machServiceName = "com.trustready.agent.helper"
    public static let helperLabel = "com.trustready.agent.helper"
    public static let clientBundleID = "com.trustready.agent.url-handler"
    public static let agentIdentifier = "com.trustready.agent"
    public static let agentExecutablePath = "/Library/TrustReady/trustready-agent"
    public static let defaultConfigDir = "/var/lib/trustready-agent"
    public static let enrolledMarkerPath = "/var/run/trustready-agent/enrolled"

    public static let helperVersion = TrustReadyAgentHelperVersion.value
}
