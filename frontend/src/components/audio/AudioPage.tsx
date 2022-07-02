import { Component } from 'react';
import { getSoundcloudUrls, SoundcloudUrl } from "../../services/andrewwillette";
import { AudioPlayer } from "./AudioPlayer";
import "./audio.css";

export class AudioPage extends Component<any, any> {
	constructor(props: any) {
		super(props);
		this.state = { soundcloudUrls: [] }
	}

	componentDidMount() {
		getSoundcloudUrls().then(soundcloudUrls => {
			let soundcloudUrlsArray = soundcloudUrls.parsedBody
			if (soundcloudUrlsArray !== undefined && soundcloudUrlsArray !== null) {
				soundcloudUrlsArray.sort(
					(v1, v2) => {
						if (v1.uiOrder > v2.uiOrder) {
							return 1;
						} else if (v1.uiOrder < v2.uiOrder) {
							return -1;
						} else {
							return 0;
						}
					})
			}
			this.setState({ soundcloudUrls: soundcloudUrlsArray })
		});
	}

	renderAudioPlayers(soundcloudUrls: SoundcloudUrl[]) {
		if (soundcloudUrls === null) {
			return <></>;
		}
		return (
			<>
				{soundcloudUrls.map((data) => {
					return <AudioPlayer key={data.url} soundcloudUrl={data.url} />
				})}
			</>
		)
	}

	render() {
		const { soundcloudUrls } = this.state;
		return (
			<div>
				{this.renderAudioPlayers(soundcloudUrls)}
			</div>
		);
	}
}
