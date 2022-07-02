import React, {Component} from "react";

export class UnauthorizedBanner extends Component<any, any> {
    render() {
        return (
            <div id={"unauthorizedBanner"}>
                Unauthorized: {this.props.unauthorizedReason}
            </div>
        );
    }
}
