import React, { Component } from 'react';
import { deleteSoundcloudUrl, getSoundcloudUrls, addSoundcloudUrl, SoundcloudUrl, login, updateSoundcloudUrls } from "../../services/andrewwillette";
import { setBearerToken } from "../../persistence/localstorage";
import { UnauthorizedBanner } from "./UnauthorizedBanner";
import { LoginSuccessBanner } from "./LoginSuccessBanner";

export class AdminPage extends Component<any, any> {
	constructor(props: any) {
		super(props);
		this.state = { soundcloudUrls: [], unauthorizedReason: null, loginSuccess: false }

		this.sendLogin = this.sendLogin.bind(this)
		this.updateSoundcloudUrlOrder = this.updateSoundcloudUrlOrder.bind(this)
		this.addSoundcloudUrl = this.addSoundcloudUrl.bind(this)
		this.saveSoundcloudUrls = this.saveSoundcloudUrls.bind(this)
	}

	componentDidMount() {
		this.updateSoundcloudUrls();
	}

	updateSoundcloudUrls() {
		getSoundcloudUrls().then(soundcloudUrls => {
			this.setState({ soundcloudUrls: soundcloudUrls.parsedBody });
		});
	}

	deleteSoundcloudUrl(soundcloudUrl: string) {
		deleteSoundcloudUrl(soundcloudUrl).then(result => {
			if (result.status === 201 || result.status === 200) {
				this.setState({ unauthorizedReason: null });
			} else {
				this.setState({ unauthorizedReason: "Not logged in, cannot delete URLS" });
			}
			this.updateSoundcloudUrls();
		});
	}

	addSoundcloudUrl() {
		const soundcloudUrl = (document.getElementById("addSoundCloudUrlInput") as HTMLInputElement).value;
		addSoundcloudUrl(soundcloudUrl).then(result => {
			if (result.status === 201 || result.status === 200) {
				this.setState({ unauthorizedReason: null });
			} else {
				this.setState({ unauthorizedReason: "Not logged in, cannot add soundcloud Url" });
			}
			this.updateSoundcloudUrls();
		});
	}

	async sendLogin() {
		let username = (document.getElementById("username") as HTMLInputElement).value
		let password = (document.getElementById("password") as HTMLInputElement).value

		let responsePromise = login(username, password)
		responsePromise.then(response => {
			if (response.status === 200) {
				const token = response.parsedBody
				if (token) {
					setBearerToken(String(token))
					this.setState({ unauthorizedReason: null, loginSuccess: true })
				}
			} else {
				this.setState({ unauthorizedReason: "Login Failed", loginSuccess: false });
			}
		});
	}

	renderAdminBanner(unauthorizedReason: string, loginSuccess: boolean) {
		if (unauthorizedReason !== null) {
			return <UnauthorizedBanner unauthorizedReason={unauthorizedReason} />
		} else if (loginSuccess) {
			return <LoginSuccessBanner />
		} else {
			return <></>
		}
	}

	updateSoundcloudUrlOrder(e: React.FormEvent<HTMLInputElement>, url: string) {
		const currentUrls: SoundcloudUrl[] = this.state.soundcloudUrls;
		const newUrls = currentUrls.map((scUrlFromMap: SoundcloudUrl) => {
			if (scUrlFromMap.url === url) {
				scUrlFromMap.uiOrder = +e.currentTarget.value
			}
			return scUrlFromMap
		})
		this.setState({ soundcloudUrls: newUrls })
	}

	// next up get this working
	saveSoundcloudUrls() {
		updateSoundcloudUrls(this.state.soundcloudUrls).then(r => {
			console.log("update service call returned to AdminPage handler then clause")
		})
		console.log(this.state.soundcloudUrls)
	}

	renderAudioManagementList(soundcloudUrls: SoundcloudUrl[]) {
		if (soundcloudUrls === null) {
			return <></>;
		}
		return (
			<>
				{soundcloudUrls.map((data) => {
					return (
						<div key={data.url} className="adminSoundcloudUrlEditbox">
							<p>{data.url}</p>
							<label htmlFor={`${data.url}-id`}>UiOrder</label>
							<input type="number"
								id={`${data.url}-id`}
								className="uiOrderInput"
								onChange={(event) => this.updateSoundcloudUrlOrder(event, data.url)}
								value={data.uiOrder} />
							<br />
							<button key={data.url}
								onClick={() => this.deleteSoundcloudUrl(data.url)}>
								Delete URL
							</button>
						</div>
					)
				})}
			</>
		)
	}

	render() {
		const { soundcloudUrls, unauthorizedReason, loginSuccess } = this.state;
		return (
			<div id="adminPage">
				<div>
					{this.renderAdminBanner(unauthorizedReason, loginSuccess)}
				</div>
				<div>
					<label htmlFor={"username"}>Username</label>
					<input id={"username"} type={"text"} />
					<br />
					<label htmlFor={"password"}>Password</label>
					<input id={"password"} type={"text"} />
					<br />
					<button onClick={this.sendLogin}>Login</button>
					<br />
					<input type={"text"} id={"addSoundCloudUrlInput"} />
					<button onClick={this.addSoundcloudUrl}>Add URL</button>
				</div>
				{this.renderAudioManagementList(soundcloudUrls)}
				<button onClick={this.saveSoundcloudUrls}>Save UiOrder</button>
			</div>
		);
	}
}
