export { setBearerToken, getBearerToken }

const bearerTokenKeyValue = "willette_bearer_token";

function setBearerToken(bearerToken: string) {
	localStorage.setItem(bearerTokenKeyValue, bearerToken);
}

function getBearerToken(): string {
	const token = localStorage.getItem(bearerTokenKeyValue);
	if (token) {
		return token;
	} else {
		return "";
	}
}
