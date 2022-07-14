import { Component } from 'react';
import "./resume.css";
import resumePdf from "./resume.pdf";

export class ResumePage extends Component<any, any> {
	render() {
		return (
			<object aria-label="Personal Resume" data={resumePdf} type="application/pdf" width="100%" height="1000em" />
		);
	}
}
