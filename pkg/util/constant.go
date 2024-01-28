package util

const WebRoot = "www"
const IndexFile = "index.html"
const Software = "Jinx"
const NotFound = "404.html"

const UnixRoot = "/usr/local"

type ServerMode string

const HTTP_SERVER ServerMode = "http_server"
const REVERSE_PROXY ServerMode = "reverse_proxy"
const FORWARD_PROXY ServerMode = "forward_proxy"
const LOAD_BALANCER ServerMode = "load_balancer"
const FTP_SERVER ServerMode = "ftp_server"

type LoadBalancerAlgo string

// ROUND_ROBIN The simplest form of load balancing, where requests are distributed sequentially to the list of servers in rotation.
// This method does not account for the current load or capacity of the servers.
const ROUND_ROBIN LoadBalancerAlgo = "round_robin"

// LEAST_CONNECTIONS This algorithm directs traffic to the server with the fewest active connections.
// It's more intelligent than round-robin because it considers the current load on each server, making it suitable for situations where session persistence is important.
const LEAST_CONNECTIONS LoadBalancerAlgo = "least_connections"

// LEAST_RESPONSE_TIME Directs traffic to the server that responds the fastest and has the fewest active connections.
// This method considers both the connection count and the server response time, offering a balance between load distribution and performance
const LEAST_RESPONSE_TIME LoadBalancerAlgo = "least_response_time"

// HASHING Various hashing methods can be used, including IP hash, where requests from the same client IP address are directed to the same server as long as the server pool remains unchanged.
// This method ensures session persistence without the need for session cookies.
const HASHING LoadBalancerAlgo = "hashing"

// WEIGHTED_ROUND_ROBIN in this algorithm the amount of request received by each server is proportional to it's weight
const WEIGHTED_ROUND_ROBIN LoadBalancerAlgo = "weighted_round_robin"

// WEIGHTED_LEAST_CONNECTIONS in this algorithm the amount of request received by each server is proportional to it's weight
const WEIGHTED_LEAST_CONNECTIONS LoadBalancerAlgo = "weighted_least_connections"

// WEIGHTED_LEAST_RESPONSE_TIME in this algorithm the amount of request received by each server is proportional to it's weight
const WEIGHTED_LEAST_RESPONSE_TIME LoadBalancerAlgo = "weighted_least_response_time"

// RANDOM Requests are distributed randomly among the servers.
// This method is less commonly used but can be effective in distributing load in a very simple and straightforward manner.
const RANDOM LoadBalancerAlgo = "random"

// RESOURCE_BASED Decisions on where to route traffic are made based on the actual resource usage of a server, such as CPU or memory utilization, ensuring that servers with sufficient resources are prioritized.
const RESOURCE_BASED LoadBalancerAlgo = "resource_based"

// GEOGRAPHICAL Traffic is distributed based on the geographical location of the client, directing clients to the server closest to them.
// This can significantly reduce latency and improve user experience for geographically distributed applications
const GEOGRAPHICAL LoadBalancerAlgo = "geographical"

const START string = "start"
const STOP string = "stop"
const RESTART string = "restart"
const DESTROY string = "destroy"
