import * as React from 'react';
import {
  ActionButton,
  EffectDiv,
  InfoItemRow,
  ThemeDiv,
  Tooltip,  
  WaitFor,
} from 'argo-ui/v2';
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

  const { metadata: { annotations, name }, spec, status } = props.resource;
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
            <ThemeDiv className='rollout__info__section'>
              <InfoItemRow
                items={{
                  content: location,
                  icon: 'fa-map-marker'
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
                  content: spec.distributionPolicy.targetShape,
                  icon: 'fa-balance-scale'
                }}
                label='Distribution Policy'
              />
              <InfoItemRow
                items={{
                  content: spec.updatePolicy.instanceRedistributionType,
                }}
                label='Update Policy'
              />
              <InfoItemRow
                items={{
                  content: spec.updatePolicy.maxSurge.fixed,
                  icon: 'fa-balance-scale'
                }}
                label='Maximum Surge'
              />
              <InfoItemRow
                items={{
                  content: spec.updatePolicy.maxUnavailable.fixed,
                  icon: 'fa-balance-scale'
                }}
                label='Maximum Unavailable'
              />
              <InfoItemRow
                items={{
                  content: spec.updatePolicy.minimalAction,
                }}
                label='Minimal Action'
              />
              <InfoItemRow
                items={{
                  content: spec.updatePolicy.replacementMethod,
                }}
                label='Replace Method'
              />                                                                
              {' '}
            </ThemeDiv>
        </div>
        <ThemeDiv className='info rollout__info'>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            height: '2em'
          }}>
            <ThemeDiv className='info__title' style={{marginBottom: '0'}}>
                Status
            </ThemeDiv>
          </div>
          <div style={{
            margin: '1em 0',
            whiteSpace: 'nowrap'
          }}>
            <InfoItemRow
              items={{
                content: !status.status.isStable ? 'Updating' : 'Stable',
              }}
              label='MIG Status'
            />
            <InfoItemRow
              items={{
                content: status.currentActions.none ? "None" : "Updating",
              }}
              label='Current Action'
            />
            <InfoItemRow
              items={{
                content: instanceGroup ? instanceGroup.managedInstances.length.toString() : "0",
              }}
              label='Instances Count'
            />            
          </div>
        </ThemeDiv>
      </div>
      <div className='rollout__row rollout__row--bottom'>
        <ThemeDiv className='info rollout__info rollout__revisions'>
          <div className='info__title'>
            Revision
          </div>
          <div style={{ marginTop: '1em' }}>
            <EffectDiv
              key={"1"}
              className='revision'
            >
              <ThemeDiv className='revision__header'>
                Revision 1
                <div style={{marginLeft: 'auto', display: 'flex', alignItems: 'center'}}>
                  <ActionButton
                      action={() => {
                        console.log("Deploy: ", instanceGroup ? instanceGroup.instanceGroupTemplate : "None")
                      }}
                      label='DEPLOY'
                      icon='fa-undo-alt'
                      style={{fontSize: '13px'}}
                      indicateLoading
                      shouldConfirm
                  />
                  <ThemeDiv className='revision__header__button' onClick={() => setCollapsed(!collapsed)}>
                    <i className={`fa ${collapsed ? 'fa-chevron-circle-down' : 'fa-chevron-circle-up'}`} />
                  </ThemeDiv>
                </div>
              </ThemeDiv>
              <ThemeDiv className='revision__images'>
                <div key={"rsInfo.objectMeta.uid"} style={{marginBottom: '1em'}}>
                  <ThemeDiv className='pods'>
                    <ThemeDiv className='pods__header'>
                      <div style={{marginRight: 'auto', flexShrink: 0}}>
                        {instanceGroup && instanceGroup.instanceGroupTemplate}
                      </div>
                    </ThemeDiv>
                    <ThemeDiv className='pods__container'>
                      <WaitFor loading={(instanceGroup && instanceGroup.managedInstances || []).length < 1}>
                        {
                          instanceGroup && instanceGroup.managedInstances.map((instance: any) => {
                            return (
                              <Tooltip content={
                                <div>
                                  <div>Name: {instance.instance}</div>
                                  <div>Status: {instance.status}</div>
                                  <div>Zone: {instance.zone}</div>
                                </div>
                              }>
                                <ThemeDiv className={`pod-icon pod-icon--success`}>
                                  <i className={`fa ${instance.status === 'RUNNING' ? 'fa-check-circle' : 'fa-exclamation-triangle'}`} />
                                </ThemeDiv>
                              </Tooltip>
                            );
                          })
                        }
                      </WaitFor>                      
                    </ThemeDiv>
                  </ThemeDiv>
                </div>
              </ThemeDiv>
            </EffectDiv>
          </div>
        </ThemeDiv>
      </div>
    </>
  );
}

export const component = Extension;