import {default as classNames} from 'classnames';
import * as React from 'react';

export interface PopupProps {
    icon?: { name: string; color: string; };
    titleColor?: string;
    title: string | React.ReactNode;
    children: React.ReactNode;    
    onClose: () => void;
    onSubmit: () => void;
}

require('./popup.scss');

const Footer = ({onClose, onSubmit}: {onClose: () => void, onSubmit: () => void}) => {
    return (
        <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'right',
            textAlign: 'right'
        }}>
            <button className='argo-button argo-button--base-o' onClick={() => onSubmit()}>
                Submit
            </button>            
            <button className='argo-button argo-button--base-o' onClick={() => onClose()}>
                Close
            </button>
        </div>
    );
};


export const Popup = (props: PopupProps) => (
    <div className='popup-overlay'>
        <div className='popup-container'>
            <div className={`row popup-container__header ${props.titleColor !== undefined ? 'popup-container__header__' + props.titleColor : 'popup-container__header__normal'}`}>
                {props.title}
            </div>
            <div className='row popup-container__body'>
                {props.icon &&
                    <div className='columns popup-container__icon'>
                        <i className={`${props.icon.name} ${props.icon.color}`}/>
                    </div>
                }
                <div className={classNames('columns', {'large-10': !!props.icon, 'large-12': !props.icon}, !props.icon && 'popup-container__body__hasNoIcon')}>
                    {props.children}
                </div>
            </div>

            <div className={classNames('row popup-container__footer', {'popup-container__footer--additional-padding': !!props.icon})}>
                <Footer onSubmit={props.onSubmit} onClose={props.onClose}/>
            </div>
        </div>
    </div>
);
