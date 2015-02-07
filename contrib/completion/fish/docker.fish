# docker.fish - docker completions for fish shell
#
# This file is generated by gen_docker_fish_completions.py from:
# https://github.com/barnybug/docker-fish-completion
#
# To install the completions:
# mkdir -p ~/.config/fish/completions
# cp docker.fish ~/.config/fish/completions
#
# Completion supported:
# - parameters
# - commands
# - containers
# - images
# - repositories

function __fish_docker_no_subcommand --description 'Test if docker has yet to be given the subcommand'
    for i in (commandline -opc)
        if contains -- $i attach build commit cp create diff events exec export history images import info insert inspect kill load login logout logs pause port ps pull push restart rm rmi run save search start stop tag top unpause version wait
            return 1
        end
    end
    return 0
end

function __fish_print_docker_containers --description 'Print a list of docker containers' -a select
    switch $select
        case running
            docker ps -a --no-trunc | command awk 'NR>1' | command awk 'BEGIN {FS="  +"}; $5 ~ "^Up" {print $1 "\n" $(NF-1)}' | tr ',' '\n'
        case stopped
            docker ps -a --no-trunc | command awk 'NR>1' | command awk 'BEGIN {FS="  +"}; $5 ~ "^Exit" {print $1 "\n" $(NF-1)}' | tr ',' '\n'
        case all
            docker ps -a --no-trunc | command awk 'NR>1' | command awk 'BEGIN {FS="  +"}; {print $1 "\n" $(NF-1)}' | tr ',' '\n'
    end
end

function __fish_print_docker_images --description 'Print a list of docker images'
    docker images | command awk 'NR>1' | command grep -v '<none>' | command awk '{print $1":"$2}'
end

function __fish_print_docker_repositories --description 'Print a list of docker repositories'
    docker images | command awk 'NR>1' | command grep -v '<none>' | command awk '{print $1}' | command sort | command uniq
end

# common options
complete -c docker -f -n '__fish_docker_no_subcommand' -l api-enable-cors -d 'Enable CORS headers in the remote API'
complete -c docker -f -n '__fish_docker_no_subcommand' -s b -l bridge -d 'Attach containers to a pre-existing network bridge'
complete -c docker -f -n '__fish_docker_no_subcommand' -l bip -d "Use this CIDR notation address for the network bridge's IP, not compatible with -b"
complete -c docker -f -n '__fish_docker_no_subcommand' -s D -l debug -d 'Enable debug mode'
complete -c docker -f -n '__fish_docker_no_subcommand' -s d -l daemon -d 'Enable daemon mode'
complete -c docker -f -n '__fish_docker_no_subcommand' -l dns -d 'Force Docker to use specific DNS servers'
complete -c docker -f -n '__fish_docker_no_subcommand' -l dns-search -d 'Force Docker to use specific DNS search domains'
complete -c docker -f -n '__fish_docker_no_subcommand' -s e -l exec-driver -d 'Force the Docker runtime to use a specific exec driver'
complete -c docker -f -n '__fish_docker_no_subcommand' -l fixed-cidr -d 'IPv4 subnet for fixed IPs (e.g. 10.20.0.0/16)'
complete -c docker -f -n '__fish_docker_no_subcommand' -l fixed-cidr-v6 -d 'IPv6 subnet for fixed IPs (e.g.: 2001:a02b/48)'
complete -c docker -f -n '__fish_docker_no_subcommand' -s G -l group -d 'Group to assign the unix socket specified by -H when running in daemon mode'
complete -c docker -f -n '__fish_docker_no_subcommand' -s g -l graph -d 'Path to use as the root of the Docker runtime'
complete -c docker -f -n '__fish_docker_no_subcommand' -s H -l host -d 'The socket(s) to bind to in daemon mode or connect to in client mode, specified using one or more tcp://host:port, unix:///path/to/socket, fd://* or fd://socketfd.'
complete -c docker -f -n '__fish_docker_no_subcommand' -s h -l help -d 'Print usage'
complete -c docker -f -n '__fish_docker_no_subcommand' -l icc -d 'Allow unrestricted inter-container and Docker daemon host communication'
complete -c docker -f -n '__fish_docker_no_subcommand' -l insecure-registry -d 'Enable insecure communication with specified registries (no certificate verification for HTTPS and enable HTTP fallback) (e.g., localhost:5000 or 10.20.0.0/16)'
complete -c docker -f -n '__fish_docker_no_subcommand' -l ip -d 'Default IP address to use when binding container ports'
complete -c docker -f -n '__fish_docker_no_subcommand' -l ip-forward -d 'Enable net.ipv4.ip_forward and IPv6 forwarding if --fixed-cidr-v6 is defined. IPv6 forwarding may interfere with your existing IPv6 configuration when using Router Advertisement.'
complete -c docker -f -n '__fish_docker_no_subcommand' -l ip-masq -d "Enable IP masquerading for bridge's IP range"
complete -c docker -f -n '__fish_docker_no_subcommand' -l iptables -d "Enable Docker's addition of iptables rules"
complete -c docker -f -n '__fish_docker_no_subcommand' -l ipv6 -d 'Enable IPv6 networking'
complete -c docker -f -n '__fish_docker_no_subcommand' -s l -l log-level -d 'Set the logging level (debug, info, warn, error, fatal)'
complete -c docker -f -n '__fish_docker_no_subcommand' -l label -d 'Set key=value labels to the daemon (displayed in `docker info`)'
complete -c docker -f -n '__fish_docker_no_subcommand' -l mtu -d 'Set the containers network MTU'
complete -c docker -f -n '__fish_docker_no_subcommand' -s p -l pidfile -d 'Path to use for daemon PID file'
complete -c docker -f -n '__fish_docker_no_subcommand' -l registry-mirror -d 'Specify a preferred Docker registry mirror'
complete -c docker -f -n '__fish_docker_no_subcommand' -s s -l storage-driver -d 'Force the Docker runtime to use a specific storage driver'
complete -c docker -f -n '__fish_docker_no_subcommand' -l selinux-enabled -d 'Enable selinux support. SELinux does not presently support the BTRFS storage driver'
complete -c docker -f -n '__fish_docker_no_subcommand' -l storage-opt -d 'Set storage driver options'
complete -c docker -f -n '__fish_docker_no_subcommand' -l tls -d 'Use TLS; implied by --tlsverify flag'
complete -c docker -f -n '__fish_docker_no_subcommand' -l tlscacert -d 'Trust only remotes providing a certificate signed by the CA given here'
complete -c docker -f -n '__fish_docker_no_subcommand' -l tlscert -d 'Path to TLS certificate file'
complete -c docker -f -n '__fish_docker_no_subcommand' -l tlskey -d 'Path to TLS key file'
complete -c docker -f -n '__fish_docker_no_subcommand' -l tlsverify -d 'Use TLS and verify the remote (daemon: verify client, client: verify daemon)'
complete -c docker -f -n '__fish_docker_no_subcommand' -s v -l version -d 'Print version information and quit'

# subcommands
# attach
complete -c docker -f -n '__fish_docker_no_subcommand' -a attach -d 'Attach to a running container'
complete -c docker -A -f -n '__fish_seen_subcommand_from attach' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from attach' -l no-stdin -d 'Do not attach STDIN'
complete -c docker -A -f -n '__fish_seen_subcommand_from attach' -l sig-proxy -d 'Proxy all received signals to the process (non-TTY mode only). SIGCHLD, SIGKILL, and SIGSTOP are not proxied.'
complete -c docker -A -f -n '__fish_seen_subcommand_from attach' -a '(__fish_print_docker_containers running)' -d "Container"

# build
complete -c docker -f -n '__fish_docker_no_subcommand' -a build -d 'Build an image from a Dockerfile'
complete -c docker -A -f -n '__fish_seen_subcommand_from build' -s f -l file -d "Name of the Dockerfile(Default is 'Dockerfile' at context root)"
complete -c docker -A -f -n '__fish_seen_subcommand_from build' -l force-rm -d 'Always remove intermediate containers, even after unsuccessful builds'
complete -c docker -A -f -n '__fish_seen_subcommand_from build' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from build' -l no-cache -d 'Do not use cache when building the image'
complete -c docker -A -f -n '__fish_seen_subcommand_from build' -l pull -d 'Always attempt to pull a newer version of the image'
complete -c docker -A -f -n '__fish_seen_subcommand_from build' -s q -l quiet -d 'Suppress the verbose output generated by the containers'
complete -c docker -A -f -n '__fish_seen_subcommand_from build' -l rm -d 'Remove intermediate containers after a successful build'
complete -c docker -A -f -n '__fish_seen_subcommand_from build' -s t -l tag -d 'Repository name (and optionally a tag) to be applied to the resulting image in case of success'

# commit
complete -c docker -f -n '__fish_docker_no_subcommand' -a commit -d "Create a new image from a container's changes"
complete -c docker -A -f -n '__fish_seen_subcommand_from commit' -s a -l author -d 'Author (e.g., "John Hannibal Smith <hannibal@a-team.com>")'
complete -c docker -A -f -n '__fish_seen_subcommand_from commit' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from commit' -s m -l message -d 'Commit message'
complete -c docker -A -f -n '__fish_seen_subcommand_from commit' -s p -l pause -d 'Pause container during commit'
complete -c docker -A -f -n '__fish_seen_subcommand_from commit' -a '(__fish_print_docker_containers all)' -d "Container"

# cp
complete -c docker -f -n '__fish_docker_no_subcommand' -a cp -d "Copy files/folders from a container's filesystem to the host path"
complete -c docker -A -f -n '__fish_seen_subcommand_from cp' -l help -d 'Print usage'

# create
complete -c docker -f -n '__fish_docker_no_subcommand' -a create -d 'Create a new container'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s a -l attach -d 'Attach to STDIN, STDOUT or STDERR.'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l add-host -d 'Add a custom host-to-IP mapping (host:ip)'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s c -l cpu-shares -d 'CPU shares (relative weight)'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l cap-add -d 'Add Linux capabilities'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l cap-drop -d 'Drop Linux capabilities'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l cidfile -d 'Write the container ID to the file'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l cpuset -d 'CPUs in which to allow execution (0-3, 0,1)'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l device -d 'Add a host device to the container (e.g. --device=/dev/sdc:/dev/xvdc:rwm)'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l dns -d 'Set custom DNS servers'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l dns-search -d "Set custom DNS search domains (Use --dns-search=. if you don't wish to set the search domain)"
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s e -l env -d 'Set environment variables'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l entrypoint -d 'Overwrite the default ENTRYPOINT of the image'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l env-file -d 'Read in a line delimited file of environment variables'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l expose -d 'Expose a port or a range of ports (e.g. --expose=3300-3310) from the container without publishing it to your host'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s h -l hostname -d 'Container host name'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s i -l interactive -d 'Keep STDIN open even if not attached'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l ipc -d 'Default is to create a private IPC namespace (POSIX SysV IPC) for the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l link -d 'Add link to another container in the form of <name|id>:alias'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l lxc-conf -d '(lxc exec-driver only) Add custom lxc options --lxc-conf="lxc.cgroup.cpuset.cpus = 0,1"'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s m -l memory -d 'Memory limit (format: <number><optional unit>, where unit = b, k, m or g)'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l mac-address -d 'Container MAC address (e.g. 92:d0:c6:0a:29:33)'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l memory-swap -d "Total memory usage (memory + swap), set '-1' to disable swap (format: <number><optional unit>, where unit = b, k, m or g)"
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l name -d 'Assign a name to the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l net -d 'Set the Network mode for the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s P -l publish-all -d 'Publish all exposed ports to random ports on the host interfaces'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s p -l publish -d "Publish a container's port to the host"
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l pid -d 'Default is to create a private PID namespace for the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l privileged -d 'Give extended privileges to this container'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l read-only -d "Mount the container's root filesystem as read only"
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l restart -d 'Restart policy to apply when a container exits (no, on-failure[:max-retry], always)'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l security-opt -d 'Security Options'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s t -l tty -d 'Allocate a pseudo-TTY'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s u -l user -d 'Username or UID'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s v -l volume -d 'Bind mount a volume (e.g., from the host: -v /host:/container, from Docker: -v /container)'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -l volumes-from -d 'Mount volumes from the specified container(s)'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -s w -l workdir -d 'Working directory inside the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from create' -a '(__fish_print_docker_images)' -d "Image"

# diff
complete -c docker -f -n '__fish_docker_no_subcommand' -a diff -d "Inspect changes on a container's filesystem"
complete -c docker -A -f -n '__fish_seen_subcommand_from diff' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from diff' -a '(__fish_print_docker_containers all)' -d "Container"

# events
complete -c docker -f -n '__fish_docker_no_subcommand' -a events -d 'Get real time events from the server'
complete -c docker -A -f -n '__fish_seen_subcommand_from events' -s f -l filter -d "Provide filter values (i.e., 'event=stop')"
complete -c docker -A -f -n '__fish_seen_subcommand_from events' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from events' -l since -d 'Show all events created since timestamp'
complete -c docker -A -f -n '__fish_seen_subcommand_from events' -l until -d 'Stream events until this timestamp'

# exec
complete -c docker -f -n '__fish_docker_no_subcommand' -a exec -d 'Run a command in a running container'
complete -c docker -A -f -n '__fish_seen_subcommand_from exec' -s d -l detach -d 'Detached mode: run command in the background'
complete -c docker -A -f -n '__fish_seen_subcommand_from exec' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from exec' -s i -l interactive -d 'Keep STDIN open even if not attached'
complete -c docker -A -f -n '__fish_seen_subcommand_from exec' -s t -l tty -d 'Allocate a pseudo-TTY'
complete -c docker -A -f -n '__fish_seen_subcommand_from exec' -a '(__fish_print_docker_containers running)' -d "Container"

# export
complete -c docker -f -n '__fish_docker_no_subcommand' -a export -d 'Stream the contents of a container as a tar archive'
complete -c docker -A -f -n '__fish_seen_subcommand_from export' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from export' -a '(__fish_print_docker_containers all)' -d "Container"

# history
complete -c docker -f -n '__fish_docker_no_subcommand' -a history -d 'Show the history of an image'
complete -c docker -A -f -n '__fish_seen_subcommand_from history' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from history' -l no-trunc -d "Don't truncate output"
complete -c docker -A -f -n '__fish_seen_subcommand_from history' -s q -l quiet -d 'Only show numeric IDs'
complete -c docker -A -f -n '__fish_seen_subcommand_from history' -a '(__fish_print_docker_images)' -d "Image"

# images
complete -c docker -f -n '__fish_docker_no_subcommand' -a images -d 'List images'
complete -c docker -A -f -n '__fish_seen_subcommand_from images' -s a -l all -d 'Show all images (by default filter out the intermediate image layers)'
complete -c docker -A -f -n '__fish_seen_subcommand_from images' -s f -l filter -d "Provide filter values (i.e., 'dangling=true')"
complete -c docker -A -f -n '__fish_seen_subcommand_from images' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from images' -l no-trunc -d "Don't truncate output"
complete -c docker -A -f -n '__fish_seen_subcommand_from images' -s q -l quiet -d 'Only show numeric IDs'
complete -c docker -A -f -n '__fish_seen_subcommand_from images' -a '(__fish_print_docker_repositories)' -d "Repository"

# import
complete -c docker -f -n '__fish_docker_no_subcommand' -a import -d 'Create a new filesystem image from the contents of a tarball'
complete -c docker -A -f -n '__fish_seen_subcommand_from import' -l help -d 'Print usage'

# info
complete -c docker -f -n '__fish_docker_no_subcommand' -a info -d 'Display system-wide information'

# inspect
complete -c docker -f -n '__fish_docker_no_subcommand' -a inspect -d 'Return low-level information on a container or image'
complete -c docker -A -f -n '__fish_seen_subcommand_from inspect' -s f -l format -d 'Format the output using the given go template.'
complete -c docker -A -f -n '__fish_seen_subcommand_from inspect' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from inspect' -a '(__fish_print_docker_images)' -d "Image"
complete -c docker -A -f -n '__fish_seen_subcommand_from inspect' -a '(__fish_print_docker_containers all)' -d "Container"

# kill
complete -c docker -f -n '__fish_docker_no_subcommand' -a kill -d 'Kill a running container'
complete -c docker -A -f -n '__fish_seen_subcommand_from kill' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from kill' -s s -l signal -d 'Signal to send to the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from kill' -a '(__fish_print_docker_containers running)' -d "Container"

# load
complete -c docker -f -n '__fish_docker_no_subcommand' -a load -d 'Load an image from a tar archive'
complete -c docker -A -f -n '__fish_seen_subcommand_from load' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from load' -s i -l input -d 'Read from a tar archive file, instead of STDIN'

# login
complete -c docker -f -n '__fish_docker_no_subcommand' -a login -d 'Register or log in to a Docker registry server'
complete -c docker -A -f -n '__fish_seen_subcommand_from login' -s e -l email -d 'Email'
complete -c docker -A -f -n '__fish_seen_subcommand_from login' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from login' -s p -l password -d 'Password'
complete -c docker -A -f -n '__fish_seen_subcommand_from login' -s u -l username -d 'Username'

# logout
complete -c docker -f -n '__fish_docker_no_subcommand' -a logout -d 'Log out from a Docker registry server'

# logs
complete -c docker -f -n '__fish_docker_no_subcommand' -a logs -d 'Fetch the logs of a container'
complete -c docker -A -f -n '__fish_seen_subcommand_from logs' -s f -l follow -d 'Follow log output'
complete -c docker -A -f -n '__fish_seen_subcommand_from logs' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from logs' -s t -l timestamps -d 'Show timestamps'
complete -c docker -A -f -n '__fish_seen_subcommand_from logs' -l tail -d 'Output the specified number of lines at the end of logs (defaults to all logs)'
complete -c docker -A -f -n '__fish_seen_subcommand_from logs' -a '(__fish_print_docker_containers running)' -d "Container"

# port
complete -c docker -f -n '__fish_docker_no_subcommand' -a port -d 'Lookup the public-facing port that is NAT-ed to PRIVATE_PORT'
complete -c docker -A -f -n '__fish_seen_subcommand_from port' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from port' -a '(__fish_print_docker_containers running)' -d "Container"

# pause
complete -c docker -f -n '__fish_docker_no_subcommand' -a pause -d 'Pause all processes within a container'
complete -c docker -A -f -n '__fish_seen_subcommand_from pause' -a '(__fish_print_docker_containers running)' -d "Container"

# ps
complete -c docker -f -n '__fish_docker_no_subcommand' -a ps -d 'List containers'
complete -c docker -A -f -n '__fish_seen_subcommand_from ps' -s a -l all -d 'Show all containers. Only running containers are shown by default.'
complete -c docker -A -f -n '__fish_seen_subcommand_from ps' -l before -d 'Show only container created before Id or Name, include non-running ones.'
complete -c docker -A -f -n '__fish_seen_subcommand_from ps' -s f -l filter -d 'Provide filter values. Valid filters:'
complete -c docker -A -f -n '__fish_seen_subcommand_from ps' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from ps' -s l -l latest -d 'Show only the latest created container, include non-running ones.'
complete -c docker -A -f -n '__fish_seen_subcommand_from ps' -s n -d 'Show n last created containers, include non-running ones.'
complete -c docker -A -f -n '__fish_seen_subcommand_from ps' -l no-trunc -d "Don't truncate output"
complete -c docker -A -f -n '__fish_seen_subcommand_from ps' -s q -l quiet -d 'Only display numeric IDs'
complete -c docker -A -f -n '__fish_seen_subcommand_from ps' -s s -l size -d 'Display total file sizes'
complete -c docker -A -f -n '__fish_seen_subcommand_from ps' -l since -d 'Show only containers created since Id or Name, include non-running ones.'

# pull
complete -c docker -f -n '__fish_docker_no_subcommand' -a pull -d 'Pull an image or a repository from a Docker registry server'
complete -c docker -A -f -n '__fish_seen_subcommand_from pull' -s a -l all-tags -d 'Download all tagged images in the repository'
complete -c docker -A -f -n '__fish_seen_subcommand_from pull' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from pull' -a '(__fish_print_docker_images)' -d "Image"
complete -c docker -A -f -n '__fish_seen_subcommand_from pull' -a '(__fish_print_docker_repositories)' -d "Repository"

# push
complete -c docker -f -n '__fish_docker_no_subcommand' -a push -d 'Push an image or a repository to a Docker registry server'
complete -c docker -A -f -n '__fish_seen_subcommand_from push' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from push' -a '(__fish_print_docker_images)' -d "Image"
complete -c docker -A -f -n '__fish_seen_subcommand_from push' -a '(__fish_print_docker_repositories)' -d "Repository"

# rename
complete -c docker -f -n '__fish_docker_no_subcommand' -a rename -d 'Rename an existing container'

# restart
complete -c docker -f -n '__fish_docker_no_subcommand' -a restart -d 'Restart a running container'
complete -c docker -A -f -n '__fish_seen_subcommand_from restart' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from restart' -s t -l time -d 'Number of seconds to try to stop for before killing the container. Once killed it will then be restarted. Default is 10 seconds.'
complete -c docker -A -f -n '__fish_seen_subcommand_from restart' -a '(__fish_print_docker_containers running)' -d "Container"

# rm
complete -c docker -f -n '__fish_docker_no_subcommand' -a rm -d 'Remove one or more containers'
complete -c docker -A -f -n '__fish_seen_subcommand_from rm' -s f -l force -d 'Force the removal of a running container (uses SIGKILL)'
complete -c docker -A -f -n '__fish_seen_subcommand_from rm' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from rm' -s l -l link -d 'Remove the specified link and not the underlying container'
complete -c docker -A -f -n '__fish_seen_subcommand_from rm' -s v -l volumes -d 'Remove the volumes associated with the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from rm' -a '(__fish_print_docker_containers stopped)' -d "Container"

# rmi
complete -c docker -f -n '__fish_docker_no_subcommand' -a rmi -d 'Remove one or more images'
complete -c docker -A -f -n '__fish_seen_subcommand_from rmi' -s f -l force -d 'Force removal of the image'
complete -c docker -A -f -n '__fish_seen_subcommand_from rmi' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from rmi' -l no-prune -d 'Do not delete untagged parents'
complete -c docker -A -f -n '__fish_seen_subcommand_from rmi' -a '(__fish_print_docker_images)' -d "Image"

# run
complete -c docker -f -n '__fish_docker_no_subcommand' -a run -d 'Run a command in a new container'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s a -l attach -d 'Attach to STDIN, STDOUT or STDERR.'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l add-host -d 'Add a custom host-to-IP mapping (host:ip)'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s c -l cpu-shares -d 'CPU shares (relative weight)'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l cap-add -d 'Add Linux capabilities'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l cap-drop -d 'Drop Linux capabilities'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l cidfile -d 'Write the container ID to the file'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l cpuset -d 'CPUs in which to allow execution (0-3, 0,1)'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s d -l detach -d 'Detached mode: run the container in the background and print the new container ID'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l device -d 'Add a host device to the container (e.g. --device=/dev/sdc:/dev/xvdc:rwm)'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l dns -d 'Set custom DNS servers'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l dns-search -d "Set custom DNS search domains (Use --dns-search=. if you don't wish to set the search domain)"
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s e -l env -d 'Set environment variables'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l entrypoint -d 'Overwrite the default ENTRYPOINT of the image'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l env-file -d 'Read in a line delimited file of environment variables'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l expose -d 'Expose a port or a range of ports (e.g. --expose=3300-3310) from the container without publishing it to your host'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s h -l hostname -d 'Container host name'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s i -l interactive -d 'Keep STDIN open even if not attached'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l ipc -d 'Default is to create a private IPC namespace (POSIX SysV IPC) for the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l link -d 'Add link to another container in the form of <name|id>:alias'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l lxc-conf -d '(lxc exec-driver only) Add custom lxc options --lxc-conf="lxc.cgroup.cpuset.cpus = 0,1"'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s m -l memory -d 'Memory limit (format: <number><optional unit>, where unit = b, k, m or g)'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l mac-address -d 'Container MAC address (e.g. 92:d0:c6:0a:29:33)'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l memory-swap -d "Total memory usage (memory + swap), set '-1' to disable swap (format: <number><optional unit>, where unit = b, k, m or g)"
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l name -d 'Assign a name to the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l net -d 'Set the Network mode for the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s P -l publish-all -d 'Publish all exposed ports to random ports on the host interfaces'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s p -l publish -d "Publish a container's port to the host"
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l pid -d 'Default is to create a private PID namespace for the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l privileged -d 'Give extended privileges to this container'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l read-only -d "Mount the container's root filesystem as read only"
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l restart -d 'Restart policy to apply when a container exits (no, on-failure[:max-retry], always)'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l rm -d 'Automatically remove the container when it exits (incompatible with -d)'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l security-opt -d 'Security Options'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l sig-proxy -d 'Proxy received signals to the process (non-TTY mode only). SIGCHLD, SIGSTOP, and SIGKILL are not proxied.'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s t -l tty -d 'Allocate a pseudo-TTY'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s u -l user -d 'Username or UID'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s v -l volume -d 'Bind mount a volume (e.g., from the host: -v /host:/container, from Docker: -v /container)'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -l volumes-from -d 'Mount volumes from the specified container(s)'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -s w -l workdir -d 'Working directory inside the container'
complete -c docker -A -f -n '__fish_seen_subcommand_from run' -a '(__fish_print_docker_images)' -d "Image"

# save
complete -c docker -f -n '__fish_docker_no_subcommand' -a save -d 'Save an image to a tar archive'
complete -c docker -A -f -n '__fish_seen_subcommand_from save' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from save' -s o -l output -d 'Write to an file, instead of STDOUT'
complete -c docker -A -f -n '__fish_seen_subcommand_from save' -a '(__fish_print_docker_images)' -d "Image"

# search
complete -c docker -f -n '__fish_docker_no_subcommand' -a search -d 'Search for an image on the registry (defaults to the Docker Hub)'
complete -c docker -A -f -n '__fish_seen_subcommand_from search' -l automated -d 'Only show automated builds'
complete -c docker -A -f -n '__fish_seen_subcommand_from search' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from search' -l no-trunc -d "Don't truncate output"
complete -c docker -A -f -n '__fish_seen_subcommand_from search' -s s -l stars -d 'Only displays with at least x stars'

# start
complete -c docker -f -n '__fish_docker_no_subcommand' -a start -d 'Start a stopped container'
complete -c docker -A -f -n '__fish_seen_subcommand_from start' -s a -l attach -d "Attach container's STDOUT and STDERR and forward all signals to the process"
complete -c docker -A -f -n '__fish_seen_subcommand_from start' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from start' -s i -l interactive -d "Attach container's STDIN"
complete -c docker -A -f -n '__fish_seen_subcommand_from start' -a '(__fish_print_docker_containers stopped)' -d "Container"

# stats
complete -c docker -f -n '__fish_docker_no_subcommand' -a stats -d "Display a live stream of one or more containers' resource usage statistics"
complete -c docker -A -f -n '__fish_seen_subcommand_from stats' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from stats' -a '(__fish_print_docker_containers running)' -d "Container"

# stop
complete -c docker -f -n '__fish_docker_no_subcommand' -a stop -d 'Stop a running container'
complete -c docker -A -f -n '__fish_seen_subcommand_from stop' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from stop' -s t -l time -d 'Number of seconds to wait for the container to stop before killing it. Default is 10 seconds.'
complete -c docker -A -f -n '__fish_seen_subcommand_from stop' -a '(__fish_print_docker_containers running)' -d "Container"

# tag
complete -c docker -f -n '__fish_docker_no_subcommand' -a tag -d 'Tag an image into a repository'
complete -c docker -A -f -n '__fish_seen_subcommand_from tag' -s f -l force -d 'Force'
complete -c docker -A -f -n '__fish_seen_subcommand_from tag' -l help -d 'Print usage'

# top
complete -c docker -f -n '__fish_docker_no_subcommand' -a top -d 'Lookup the running processes of a container'
complete -c docker -A -f -n '__fish_seen_subcommand_from top' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from top' -a '(__fish_print_docker_containers running)' -d "Container"

# unpause
complete -c docker -f -n '__fish_docker_no_subcommand' -a unpause -d 'Unpause a paused container'
complete -c docker -A -f -n '__fish_seen_subcommand_from unpause' -a '(__fish_print_docker_containers running)' -d "Container"

# version
complete -c docker -f -n '__fish_docker_no_subcommand' -a version -d 'Show the Docker version information'

# wait
complete -c docker -f -n '__fish_docker_no_subcommand' -a wait -d 'Block until a container stops, then print its exit code'
complete -c docker -A -f -n '__fish_seen_subcommand_from wait' -l help -d 'Print usage'
complete -c docker -A -f -n '__fish_seen_subcommand_from wait' -a '(__fish_print_docker_containers running)' -d "Container"


