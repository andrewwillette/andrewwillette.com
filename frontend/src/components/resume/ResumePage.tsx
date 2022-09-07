import { Component } from 'react';
import "./resume.css";

export class ResumePage extends Component<any, any> {
	readonly resumeUrl = "https://andrewwillette.s3.us-east-2.amazonaws.com/newdir/resume.pdf";

	render() {
		return (
			<object aria-label="Personal Resume" data={this.resumeUrl} type="application/pdf" width="100%" height="1000em" />
		);
	}
}
