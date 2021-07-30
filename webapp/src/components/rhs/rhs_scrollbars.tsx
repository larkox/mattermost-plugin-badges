import React from 'react';

import Scrollbars from 'react-custom-scrollbars';

function renderView(props: any) {
    return (
        <div
            {...props}
            className='scrollbar--view'
        />);
}

function renderThumbHorizontal(props: any) {
    return (
        <div
            {...props}
            className='scrollbar--horizontal'
        />);
}

function renderThumbVertical(props: any) {
    return (
        <div
            {...props}
            className='scrollbar--vertical'
        />);
}

const RHSScrollbars = ({children}: {children: React.ReactNode[]}) => {
    return (
        <Scrollbars
            autoHide={true}
            autoHideTimeout={500}
            autoHideDuration={500}
            renderThumbHorizontal={renderThumbHorizontal}
            renderThumbVertical={renderThumbVertical}
            renderView={renderView}
            style={{flex: '1 1 auto', height: ''}}
        >
            {children}
        </Scrollbars>
    );
};

export default RHSScrollbars;
