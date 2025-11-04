/*
Package ensure provides checking for the presence of the netdevsim system bus
and loading the required kernel module if it isn't present and the caller is
root. Normally, the netdevsim system bus (for managing netdevsim netdevs) is
accessible at “/sys/bus/netdevsim”.
*/
package ensure
