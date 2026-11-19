package io.nekohasekai.sagernet.ui.profile

import androidx.compose.runtime.Composable
import io.nekohasekai.sagernet.database.DataStore
import io.nekohasekai.sagernet.fmt.shadowquic.ShadowQUICBean
import io.nekohasekai.sagernet.ktx.applyDefaultValues
import io.nekohasekai.sagernet.ui.compose.ShadowQUICProfileSettingsScreen

class ShadowQUICSettingsActivity : ProfileSettingsActivity<ShadowQUICBean>() {
    override val usesComposePreferences = true
    override fun createEntity() = ShadowQUICBean().applyDefaultValues()

    override fun ShadowQUICBean.init() {
        DataStore.profileName = name
        DataStore.serverAddress = serverAddress
        DataStore.serverPort = serverPort
        DataStore.serverUsername = username
        DataStore.serverPassword = password
        DataStore.serverSNI = sni
        DataStore.serverALPN = alpn
        DataStore.serverCongestionController = congestionControl
        DataStore.serverUDPRelayMode = if (udpOverStream) "stream" else "datagram"
        DataStore.serverReduceRTT = zeroRTT
    }

    override fun ShadowQUICBean.serialize() {
        name = DataStore.profileName
        serverAddress = DataStore.serverAddress
        serverPort = DataStore.serverPort
        username = DataStore.serverUsername
        password = DataStore.serverPassword
        sni = DataStore.serverSNI
        alpn = DataStore.serverALPN
        congestionControl = DataStore.serverCongestionController
        udpOverStream = DataStore.serverUDPRelayMode == "stream"
        zeroRTT = DataStore.serverReduceRTT
    }

    @Composable
    override fun ComposePreferences() = ShadowQUICProfileSettingsScreen()
}
