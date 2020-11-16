import React, { useState } from 'react';
import { SVG } from '../../../components/factorsComponents';
import SaveQuery from '../../../components/SaveQuery';
import ConfirmationModal from '../../../components/ConfirmationModal';

function ResultsHeader({
  setShowResult, requestQuery, querySaved, setQuerySaved
}) {
  const [showModal, setShowModal] = useState(false);
  const [showSaveModal, setShowSaveModal] = useState(false);

  const handleSaveQuery = () => {
    console.log('user clicked save query');
    setShowModal(false);
    setShowSaveModal(true);
  };

  const exitResultsPage = () => {
    setShowModal(false);
    setShowResult(false);
  };

  const handleBreadcrumbClick = () => {
    if (querySaved) {
      setShowResult(false);
    } else {
      setShowModal(true);
    }
  };

  let saveText = 'Unsaved';

  if (querySaved) {
    saveText = 'Saved';
  }

  return (
    <>
            <ConfirmationModal
                visible={showModal}
                confirmationText="You have not saved this query yet. If you like to run this query often, we recommend to save this."
                onOk={handleSaveQuery}
                onCancel={exitResultsPage}
                title="Exit without saving?"
                width={600}
                okText="Save query"
                cancelText="Yes, exit"
            />
            <div className="flex py-4 justify-between items-center">
                <div className="leading-4">
                    <div onClick={handleBreadcrumbClick} className="flex items-center cursor-pointer">
                        <div>
                            <SVG name={'breadcrumb'} color="#0B1E39" size={25} />
                        </div>
                        <div style={{ color: '#0E2647', opacity: 0.56, fontSize: '14px' }} className="font-bold leading-5"> / Query / {saveText}</div>
                    </div>
                </div>
                <div>
                    <SaveQuery
                        requestQuery={requestQuery}
                        setQuerySaved={setQuerySaved}
                        visible={showSaveModal}
                        setVisible={setShowSaveModal}
                    />
                </div>
            </div>
    </>

  );
}

export default ResultsHeader;
