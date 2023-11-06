const ssid = document.getElementById("ssid");
const password = document.getElementById("password");
const button = document.getElementById("submit");	

button.onclick = async () => {
	const id = ssid.value;
	const pass = password.value;
		
	const res = await fetch("http://192.168.1.4/connect", {
		body: JSON.stringify({
			id,
			pass,
		}),
		method: "POST",
	});
}
