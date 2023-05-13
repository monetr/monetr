import MButton from "components/MButton";
import MLogo from "components/MLogo";
import MSpan from "components/MSpan";
import React from "react";


export default function SidebarNew(): JSX.Element {
  return (
    <div className="w-full h-full">
      <div className="flex flex-col w-72 z-50 left-0 top-0 bottom-0 fixed bg-purple-800">
        <div className="pb-4 px-6 overflow-y-auto flex-col gap-y-5 flex-grow flex">
          <div className="items-center flex-shrink-0 h-16 flex">
            <MLogo className="h-8 w-auto" />
            <MSpan>monetr</MSpan>
          </div>
          <MButton color="primary" variant="solid" role="form" type="submit">
            Sign In
          </MButton>
        </div>
      </div>
      <div className="pl-72 flex justify-center items-center h-full">
        <span className="text-5xl">[ CONTENT ]</span>
      </div>
    </div>
  );
}
