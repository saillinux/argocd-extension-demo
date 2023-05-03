import * as React from 'react';

import {
  Menu,
  Tooltip,
  InfoItemRow,
  Input,
  Text,
} from 'argo-ui/v2';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {
  faChevronCircleUp,
  faCheck,  
} from '@fortawesome/free-solid-svg-icons';

import "./index.scss";

const EXTPATH = "/extensions/extdemo";

// Create a Header to access extension backend
const createHttpHeaders = () => {
  const myHeaders = new Headers();
  // The backend for this resides in the application argocd-extension-demo
  myHeaders.append("Argocd-Application-Name", "argocd:argocd-extension-demo")
  myHeaders.append("Argocd-Project-Name", "default")
  return myHeaders;
}

const getInstanceGroup = async (projectId: string, region: string, instanceGroupName: string) => {
  const myHeaders = createHttpHeaders();
  const myRequest = new Request(`${EXTPATH}/compute/instancegroup/${projectId}/${region}/${instanceGroupName}`, {
    method: "GET",
    headers: myHeaders,
    mode: "cors",
    cache: "default",
  });
  const response = await fetch(myRequest);
  const instanceGroup = await response.json();
  return instanceGroup;
};

export const Extension = (props: {
  tree: any;
  resource: any;
}) => {
  const [instanceGroup, setInstanceGroup] = React.useState(null);
  const [collapsed, setCollapsed] = React.useState(true);

  console.log(props);

  const { metadata: { annotations, name }, spec } = props.resource;
  const { location, targetSize } = spec;
  
  const projectId = annotations['cnrm.cloud.google.com/project-id'];
  
  React.useEffect(() => {
    if (instanceGroup) {
      return;
    }

    (async () => {
      const instanceGroup = await getInstanceGroup(projectId, location, name);
      console.log(instanceGroup);
      setInstanceGroup(instanceGroup);
    })();
    return () => {
      console.log("");
    };
  });

  return (
    <>
      <div className='rollout__row rollout__row--top'>
        <div className='info rollout__info'>
            <div className='info__title'>
              Summary
            </div>
            <InfoItemRow
                items={{
                  content: "rollout.strategy",
                  icon: "fa-time",
                }}
                label='Strategy'
            />
            <div className='rollout__info__section'>
              <React.Fragment>
                <InfoItemRow
                  items={{
                    content: location,
                    icon: 'fa-shoe-prints'
                  }}
                  label='Location'
                />
                <InfoItemRow
                  items={{
                    content: targetSize,
                    icon: 'fa-balance-scale-right'
                  }}
                  label='Target Size'
                />
                <InfoItemRow
                  items={{
                    content: 'rollout.actualWeight',
                    icon: 'fa-balance-scale'
                  }}
                  label='Actual Weight'
                />                    
                {' '}
              </React.Fragment>
            </div>
        </div>
        <div className='info rollout__info'>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            height: '2em'
          }}>
            <div className='info__title' style={{marginBottom: '0'}}>
                Containers
            </div>
          </div>
          <div style={{margin: '1em 0', whiteSpace: 'nowrap'}}>
            <div style={{marginBottom: '0.5em', fontWeight: 600, fontSize: '14px'}}>
              container.name
            </div>
            <div style={{width: '100%', height: '2em', minWidth: 0}}>
              <Input
                value={"container.image"}
                style={{
                  width: '100%',
                  cursor: 'default',
                  color: 'black'
                }}
                disabled={true}
              />
            </div>
        </div>
          <div className='containers__few'>
            <span style={{ marginRight: '5px' }}>
              <i className='fa fa-boxes' />
            </span>
          </div>
        </div>
      </div>
      <div className='rollout__row rollout__row--bottom'>
        <div className='info rollout__info rollout__revisions'>
          <div className='info__title'>
            Revisions
          </div>
          <div style={{ marginTop: '1em' }}>
            <div
              key={"1"}
              className='revision'
            >
              <div className='revision__header'>
                Revision 1
                <div style={{marginLeft: 'auto', display: 'flex', alignItems: 'center'}}>
                    <FontAwesomeIcon
                      icon={faChevronCircleUp}
                      className='revision__header__button'
                      onClick={() => setCollapsed(!collapsed)}
                    />
                </div>
              </div>
              <div className='revision__images'>
                <InfoItemRow
                  key={"img.image"}
                  label={
                    "label"
                  }
                  items={[
                    {
                      content: 'Revision Content'
                    }
                  ]}
                />
                <div key={"rsInfo.objectMeta.uid"} style={{marginBottom: '1em'}}>
                  <div className='pods'>
                    <Tooltip content={"rsName"}>
                      <div className='pods__header'>
                        <Text style={{ maxWidth: '100%' }}>
                          rsName
                        </Text>
                        <Tooltip content={"status"}>
                          <i className={`fa fa-check-circle status-icon--healthy`} />
                        </Tooltip>
                        <div style={{marginLeft: 'auto', flexShrink: 0}}>
                          Revision 1
                        </div>
                      </div>
                    </Tooltip>
                    <div className='pods__container'>
                      <Menu
                        items={[
                          "menu1"
                        ]}
                      >
                        <Tooltip content={
                          <div>
                            <div>Status: </div>
                            <div>name</div>
                          </div>
                        }>
                          <div className={`pod-icon pod-icon--success`}>
                            <FontAwesomeIcon icon={faCheck} spin={false} />
                          </div>
                        </Tooltip>
                      </Menu>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

export const component = Extension;