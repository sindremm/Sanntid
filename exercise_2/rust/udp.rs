use std::net::UdpSocket;

const SERVER_IP_ADDRESS: &'static str = "10.100.23.129";

const port: &'static str = "20005";


fn main() -> std::io::Result<()> {
    {
        let socket = UdpSocket::bind(full_address)?;
        let full_address: str = SERVER_IP_ADDRESS + ":" + port;

        // Receives a single datagram message on the socket. If `buf` is too small to hold
        // the message, it will be cut off.
        let mut buf = [0; 10];
        let (amt, src) = socket.recv_from(&mut buf)?;

        println!("{}", amt);
    } // the socket is closed here
    Ok(())
}