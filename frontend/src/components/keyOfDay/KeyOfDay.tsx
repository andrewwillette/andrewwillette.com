import { Component } from 'react';
import { getKeyOfDay } from "../../services/andrewwillette";

export class KeyOfDay extends Component<any, any> {
	constructor(props: any) {
		super(props);
		this.state = { kod: "" }
	}
	componentDidMount() {
		getKeyOfDay().then(keyOfDay => {
			let kod = keyOfDay.parsedBody
			this.setState({ kod: kod })
		});
	}

	render() {
		const { kod } = this.state;
		return (
			<div>
				{kod}
			</div>
		);
	}
}
