package io.nekohasekai.sagernet.fmt.shadowquic

import io.nekohasekai.sagernet.fmt.subscriptionBoolean
import io.nekohasekai.sagernet.ktx.linkBuilder
import io.nekohasekai.sagernet.ktx.toLink
import moe.matsuri.nb4a.SingBoxOptions
import moe.matsuri.nb4a.utils.listByLineOrComma
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import moe.matsuri.nb4a.utils.JavaUtil.gson

fun parseShadowQUIC(url: String): ShadowQUICBean {
    val link = url.replaceFirst("sq://", "https://").toHttpUrlOrNull()
        ?: error("invalid ShadowQUIC link")
    return ShadowQUICBean().apply {
        name = link.fragment.orEmpty()
        serverAddress = link.host
        serverPort = link.port.takeIf { it > 0 } ?: 443
        username = link.username.ifBlank { error("missing username") }
        password = link.password.ifBlank { error("missing password") }
        sni = link.queryParameter("sni")?.takeIf { it.isNotBlank() }
            ?: error("missing sni")
        alpn = link.queryParameter("alpn") ?: "h3"
        congestionControl = link.queryParameter("congestion_control") ?: "bbr"
        udpOverStream = when (link.queryParameter("udp_mode")) {
            "datagram" -> false
            "stream" -> true
            else -> true
        }
        zeroRTT = link.queryParameter("zero_rtt")?.subscriptionBoolean() ?: false
    }
}

fun ShadowQUICBean.toUri(): String {
    require(username.isNotBlank()) { "missing username" }
    require(password.isNotBlank()) { "missing password" }
    require(sni.isNotBlank()) { "missing sni" }
    val builder = linkBuilder().username(username).password(password)
        .host(serverAddress).port(serverPort)
    builder.addQueryParameter("sni", sni)
    builder.addQueryParameter("udp_mode", if (udpOverStream) "stream" else "datagram")
    builder.addQueryParameter("congestion_control", congestionControl)
    if (zeroRTT) builder.addQueryParameter("zero_rtt", "true")
    val protocols = alpn.listByLineOrComma()
    if (protocols.isNotEmpty()) builder.addQueryParameter("alpn", protocols.joinToString(","))
    if (name.isNotBlank()) builder.fragment(name)
    return builder.toLink("sq")
}

fun buildSingBoxOutboundShadowQUICBean(bean: ShadowQUICBean): SingBoxOptions.SingBoxOption {
    require(bean.username.isNotBlank()) { "ShadowQUIC username is required" }
    require(bean.password.isNotBlank()) { "ShadowQUIC password is required" }
    require(bean.sni.isNotBlank()) { "ShadowQUIC SNI is required" }
    val config = linkedMapOf<String, Any>(
        "type" to "shadowquic",
        "server" to bean.serverAddress,
        "server_port" to bean.serverPort,
        "username" to bean.username,
        "password" to bean.password,
        "server_name" to bean.sni,
        "congestion_control" to bean.congestionControl,
        "udp_over_stream" to bean.udpOverStream,
        "zero_rtt_handshake" to bean.zeroRTT,
    )
    val protocols = bean.alpn.listByLineOrComma()
    if (protocols.isNotEmpty()) config["alpn"] = protocols
    return SingBoxOptions.CustomSingBoxOption(gson.toJson(config))
}
