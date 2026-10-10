package com.fujiwaracreative.famloop.utils

import java.net.Inet4Address
import java.net.Inet6Address
import java.net.NetworkInterface
import java.util.Collections

object NetworkUtils {

    data class DeviceIPs(
        val ipv4: String?,
        val ipv6: String?
    )

    fun getDeviceIPAddresses(): DeviceIPs {
        var ipv4: String? = null
        var ipv6: String? = null

        try {
            val interfaces = Collections.list(NetworkInterface.getNetworkInterfaces())
            for (intf in interfaces) {
                if (intf.isLoopback || !intf.isUp) continue
                val addrs = Collections.list(intf.inetAddresses)
                for (addr in addrs) {
                    if (addr.isLoopbackAddress) continue
                    if (addr is Inet4Address && ipv4 == null) {
                        ipv4 = addr.hostAddress
                    } else if (addr is Inet6Address && ipv6 == null) {
                        val host = addr.hostAddress ?: ""
                        // Buang zone index jika ada (misal %wlan0)
                        val delim = host.indexOf('%')
                        ipv6 = if (delim < 0) host else host.substring(0, delim)
                    }
                }
            }
        } catch (_: Exception) {}

        return DeviceIPs(ipv4 = ipv4 ?: "127.0.0.1", ipv6 = ipv6)
    }
}
