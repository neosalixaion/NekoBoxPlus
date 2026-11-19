package io.nekohasekai.sagernet.fmt.shadowquic;

import androidx.annotation.NonNull;
import com.esotericsoftware.kryo.io.ByteBufferInput;
import com.esotericsoftware.kryo.io.ByteBufferOutput;
import io.nekohasekai.sagernet.fmt.AbstractBean;
import io.nekohasekai.sagernet.fmt.KryoConverters;

public class ShadowQUICBean extends AbstractBean {
    public String username;
    public String password;
    public String sni;
    public String alpn;
    public String congestionControl;
    public Boolean zeroRTT;
    public Boolean udpOverStream;

    @Override
    public void initializeDefaultValues() {
        if (serverPort == null) serverPort = 443;
        super.initializeDefaultValues();
        if (username == null) username = "";
        if (password == null) password = "";
        if (sni == null) sni = "";
        if (alpn == null) alpn = "h3";
        if (congestionControl == null) congestionControl = "bbr";
        if (zeroRTT == null) zeroRTT = false;
        if (udpOverStream == null) udpOverStream = false;
    }

    @Override
    public void serialize(ByteBufferOutput output) {
        super.serialize(output);
        output.writeInt(4);
        output.writeString(username);
        output.writeString(password);
        output.writeString(sni);
        output.writeString(alpn);
        output.writeString(congestionControl);
        output.writeBoolean(zeroRTT);
        output.writeBoolean(udpOverStream);
    }

    @Override
    public void deserialize(ByteBufferInput input) {
        super.deserialize(input);
        int version = input.readInt();
        username = input.readString();
        password = input.readString();
        sni = input.readString();
        alpn = input.readString();
        congestionControl = input.readString();
        if (version <= 3 && "new-reno".equals(congestionControl)) {
            congestionControl = "new_reno";
        }
        if (version <= 3 && "brutal".equals(congestionControl)) {
            congestionControl = "bbr";
        }
        zeroRTT = input.readBoolean();
        udpOverStream = input.readBoolean();
        if (version >= 1 && version <= 3) {
            boolean disableALPN = input.readBoolean();
            if (disableALPN) {
                alpn = "";
            }
            if (!disableALPN && alpn.isEmpty()) {
                alpn = "h3";
            }
            input.readBoolean(); // legacy useSunnyQUIC
        }
        if (version >= 2 && version <= 3) {
            input.readString(); // legacy certificate
        }
        if (version == 3) {
            input.readLong(); // legacy brutal upload bandwidth
        }
    }

    @Override
    public boolean canTCPing() {
        return false;
    }

    @NonNull
    @Override
    public ShadowQUICBean clone() {
        return KryoConverters.deserialize(new ShadowQUICBean(), KryoConverters.serialize(this));
    }

    @NonNull
    @Override
    public String getHash() {
        return buildTypedHash("shadowquic");
    }

    public static final Creator<ShadowQUICBean> CREATOR = new CREATOR<ShadowQUICBean>() {
        @NonNull @Override public ShadowQUICBean newInstance() { return new ShadowQUICBean(); }
        @Override public ShadowQUICBean[] newArray(int size) { return new ShadowQUICBean[size]; }
    };
}
