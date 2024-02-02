package constant

import "jinx/pkg/util/types"

const DEFAULT_WEBSITE_ROOT = "www"
const INDEX_FILE = "index.html"
const SOFTWARE_NAME = "Jinx"
const NOT_FOUND = "404.html"
const IMAGE_DIR = "images"
const VERSION_NUMBER = "1.0.0"

const ROOT = "/home/unix-martin"
const BASE = ROOT + "/jinx"
const CONFIG_FILE_PATH = BASE + "/" + CONFIG_FILE
const LOG_ROOT = "logs"
const DEFAULT_WEBSITE_ROOT_DIR = BASE + "/" + HTTP_SERVER + "/" + DEFAULT_WEBSITE_ROOT
const DEFAULT_IP = "127.0.0.1"
const CONFIG_FILE = "jinx_config.json"

const JINX_ICO_URL = "https://gemkox-spaces.nyc3.cdn.digitaloceanspaces.com/jinx/jinx.ico"
const JINX_SVG_URL = "https://gemkox-spaces.nyc3.cdn.digitaloceanspaces.com/jinx/jinx.svg"
const JINX_INDEX_URL = "https://gemkox-spaces.nyc3.cdn.digitaloceanspaces.com/jinx/index.html"
const JINX_404_URL = "https://gemkox-spaces.nyc3.cdn.digitaloceanspaces.com/jinx/404.html"
const JINX_CSS_URL = "https://gemkox-spaces.nyc3.cdn.digitaloceanspaces.com/jinx/style.css"

const JINX_ICO_FILE = "jinx.ico"
const JINX_SVG_FILE = "jinx.svg"
const JINX_INDEX_FILE = "index.html"
const JINX_404_FILE = "404.html"
const JINX_CSS_FILE = "style.css"

const HTTP_SERVER types.ServerMode = "http_server"
const REVERSE_PROXY types.ServerMode = "reverse_proxy_server"
const FORWARD_PROXY types.ServerMode = "forward_proxy_server"
const LOAD_BALANCER types.ServerMode = "load_balancing_server"

const VERSION = "version"

// ROUND_ROBIN The simplest form of load balancing, where requests are distributed sequentially to the list of servers in rotation.
// This method does not account for the current load or capacity of the servers.
const ROUND_ROBIN types.LoadBalancerAlgo = "round_robin"

// LEAST_CONNECTIONS This algorithm directs traffic to the server with the fewest active connections.
// It's more intelligent than round-robin because it considers the current load on each server, making it suitable for situations where session persistence is important.
const LEAST_CONNECTIONS types.LoadBalancerAlgo = "least_connections"

// LEAST_RESPONSE_TIME Directs traffic to the server that responds the fastest and has the fewest active connections.
// This method considers both the connection count and the server response time, offering a balance between load distribution and performance
const LEAST_RESPONSE_TIME types.LoadBalancerAlgo = "least_response_time"

// HASHING Various hashing methods can be used, including IP hash, where requests from the same client IP address are directed to the same server as long as the server pool remains unchanged.
// This method ensures session persistence without the need for session cookies.
const HASHING types.LoadBalancerAlgo = "hashing"

// WEIGHTED_ROUND_ROBIN in this algorithm the amount of request received by each server is proportional to it's weight
const WEIGHTED_ROUND_ROBIN types.LoadBalancerAlgo = "weighted_round_robin"

// WEIGHTED_LEAST_CONNECTIONS in this algorithm the amount of request received by each server is proportional to it's weight
const WEIGHTED_LEAST_CONNECTIONS types.LoadBalancerAlgo = "weighted_least_connections"

// WEIGHTED_LEAST_RESPONSE_TIME in this algorithm the amount of request received by each server is proportional to it's weight
const WEIGHTED_LEAST_RESPONSE_TIME types.LoadBalancerAlgo = "weighted_least_response_time"

// RANDOM Requests are distributed randomly among the servers.
// This method is less commonly used but can be effective in distributing load in a very simple and straightforward manner.
const RANDOM types.LoadBalancerAlgo = "random"

// RESOURCE_BASED Decisions on where to route traffic are made based on the actual resource usage of a server, such as CPU or memory utilization, ensuring that servers with sufficient resources are prioritized.
const RESOURCE_BASED types.LoadBalancerAlgo = "resource_based"

// GEOGRAPHICAL Traffic is distributed based on the geographical location of the client, directing clients to the server closest to them.
// This can significantly reduce latency and improve user experience for geographically distributed applications
const GEOGRAPHICAL types.LoadBalancerAlgo = "geographical"

const START string = "start"
const STOP string = "stop"
const RESTART string = "restart"
const DESTROY string = "destroy"
